package prometheus_wrapper

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"panda-go/common/log"
	"panda-go/component/kafka/consumer"
	"panda-go/component/kafka/producer"
	panda_nacos "panda-go/component/nacos/base"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	goroutineTopic = "panda_goroutine_count"
	goroutineCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: goroutineTopic,
		Help: "Number of Goroutines in every service process.",
	},
		[]string{"service_name", "pod_name"})
	initCtx = context.TODO()
)

func ClosePrometheus() {
	initCtx.Done()
}

func updateGoroutineCount(ctx context.Context) {
	cc, err := consumer.GetTopicMessageChan(ctx, goroutineTopic, true)
	if err != nil {
		log.Error(ctx, "[updateGoroutineCount] error", zap.Error(err))
		return
	}
	for {
		select {
		case msg := <-cc:
			log.Debug(ctx, "[updateGoroutineCount]", zap.ByteString("msg.key", msg.Key), zap.ByteString("msg.Value", msg.Value))
			count, err := strconv.ParseFloat(string(msg.Value), 64)
			if err != nil {
				log.Error(ctx, "[updateGoroutineCount] ParseFloat fail", zap.Error(err))
				return
			}
			key := string(msg.Key)
			keys := strings.Split(key, "/")
			if len(keys) == 2 {
				goroutineCount.WithLabelValues(keys[0], keys[1]).Set(count)
			}
		case <-ctx.Done():
			log.Info(ctx, "[updateGoroutineCount] exit ctx done", zap.Error(ctx.Err()))
			return
		}
	}
}

func InitCollectPrometheus() {
	prometheus.MustRegister(goroutineCount)
	// 定期更新 Goroutine 数量
	go updateGoroutineCount(initCtx)

}

func InitPushPrometheus() {
	go func() {
		ticker := time.Tick(10 * time.Second)
		for {
			select {
			case <-ticker:
				err := producer.SentTopicMessage(initCtx, goroutineTopic, panda_nacos.ServerName+"/"+panda_nacos.PodName, fmt.Sprintf("%d", runtime.NumGoroutine()))
				if err != nil {
					log.Error(initCtx, "[InitPushPrometheus] SentTopicMessage err", zap.Error(err))
					return
				}
			case <-initCtx.Done():
				log.Info(initCtx, "[InitPushPrometheus] initCtx done", zap.Error(initCtx.Err()))
				return
			}
		}
	}()
}
