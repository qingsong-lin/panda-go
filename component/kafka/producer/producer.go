package producer

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"panda-go/common/log"
	panda_nacos "panda-go/component/nacos/base"
	"sync"

	"github.com/IBM/sarama"
)

func (m *MockSCRAMClient) Begin(_, _, _ string) (err error) {
	return nil
}

func (m *MockSCRAMClient) Step(challenge string) (response string, err error) {
	if challenge == "" {
		return "ping", nil
	}
	if challenge == "pong" {
		m.done = true
		return "", nil
	}
	return "", errors.New("failed to authenticate :(")
}

func (m *MockSCRAMClient) Done() bool {
	return m.done
}

type MockSCRAMClient struct {
	done bool
}

var (
	once     sync.Once
	producer sarama.SyncProducer
)

func CloseKafkaProducer() {
	if producer != nil {
		producer.Close()
	}
}

func GetKafkaProducer(ctx context.Context) sarama.SyncProducer {
	once.Do(func() {
		var kafkaUrl string
		kafkaUrl, err := panda_nacos.GetNacosClient().GetConfig(ctx, &panda_nacos.ConfigParam{
			DataId: panda_nacos.ServerName + ".kafka.config",
		})
		if err != nil {
			log.Warn(ctx, "[GetKafkaClient] get current service kafka config fail", zap.Error(err))
			kafkaUrl, err = panda_nacos.GetNacosClient().GetConfig(ctx, &panda_nacos.ConfigParam{
				DataId: "common.kafka.config",
				Group:  "DEFAULT_GROUP",
			})
			if err != nil {
				log.Fatal(ctx, "[GetKafkaClient] get kafka config all fail", zap.Error(err))
			}
			if kafkaUrl == "" {
				log.Fatal(ctx, "[GetKafkaClient] get common kafka config empty", zap.Error(err))
			}
		}
		if kafkaUrl == "" {
			log.Fatal(ctx, "[GetKafkaClient] get service kafka config empty", zap.String("service_name", panda_nacos.ServerName), zap.Error(err))
		}
		config := sarama.NewConfig()
		config.Producer.Return.Successes = true
		producer, err = sarama.NewSyncProducer([]string{kafkaUrl}, config)
		if err != nil {
			log.Fatal(ctx, "[GetKafkaClient] sarama.NewConsumer error", zap.Error(err))
		}
	})
	return producer
}

func SentTopicMessage(ctx context.Context, topic string, key string, val string) error {
	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(val),
	}
	partition, offset, err := GetKafkaProducer(ctx).SendMessage(message)
	if err != nil {
		log.Error(ctx, "[SentTopicMessage] Error producing message:", zap.Error(err))
		return err
	}
	log.Debug(ctx, fmt.Sprintf("Produced message to topic %s, partition %d, offset %d\n", topic, partition, offset))
	return nil
}

//
//func main() {
//	ctx := context.TODO()
//	brokerAddress := "kafka.common.svc.cluster.local:9092" // 指定 Kafka 集群中任一 Broker 的地址和端口
//	topic := "your_topic"                                  // 指定要生产消息的主题
//	// 配置 SASL/SCRAM 认证参数
//	config := sarama.NewConfig()
//	//config.Net.SASL.Enable = true
//	//config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
//	//config.Net.SASL.User = "user1"
//	//config.Net.SASL.Password = "ozSUKGWtHZ"
//	//config.Net.SASL.Handshake = true
//	config.Producer.Return.Successes = true
//	//config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &MockSCRAMClient{} }
//	// 创建 Kafka 生产者
//	producer, err := sarama.NewSyncProducer([]string{brokerAddress}, config)
//	if err != nil {
//		log.Fatal(ctx, "Error creating producer:", zap.Error(err))
//	}
//	defer func() {
//		if err := producer.Close(); err != nil {
//			log.Fatal(ctx, "Error closing producer:", zap.Error(err))
//		}
//	}()
//
//	// 生产消息
//	message := &sarama.ProducerMessage{
//		Topic: topic,
//		Key:   sarama.StringEncoder("key"),
//		Value: sarama.StringEncoder("Hello, Kafka with SASL/SCRAM!"),
//	}
//	partition, offset, err := producer.SendMessage(message)
//	if err != nil {
//		log.Fatal(ctx, "Error producing message:", zap.Error(err))
//	}
//
//	log.Infof(ctx, "Produced message to topic %s, partition %d, offset %d\n", topic, partition, offset)
//
//	// 等待信号，例如 Ctrl+C
//	sigterm := make(chan os.Signal, 1)
//	signal.Notify(sigterm, os.Interrupt)
//	select {
//	case <-sigterm:
//		log.Info(ctx, "Interrupted. Closing producer.")
//	}
//}
