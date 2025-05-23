package consumer

import (
	"context"
	"go.uber.org/zap"
	"panda-go/common/log"
	panda_nacos "panda-go/component/nacos/base"
	"sync"

	"github.com/IBM/sarama"
)

var (
	once                 sync.Once
	onceConsumer         sync.Once
	client               sarama.Consumer
	consumerGroup        sarama.ConsumerGroup
	topic2MessageChanMap sync.Map
)

func CloseKafkaClient() {
	if client != nil {
		client.Close()
	}
	if consumerGroup != nil {
		consumerGroup.Close()
	}
}

func GetKafkaClient(ctx context.Context) sarama.Consumer {
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
		config.Consumer.Return.Errors = true
		client, err = sarama.NewConsumer([]string{kafkaUrl}, config)
		if err != nil {
			log.Fatal(ctx, "[GetKafkaClient] sarama.NewConsumer error", zap.Error(err))
		}
	})
	return client
}

func GetKafkaConsumer(ctx context.Context, isNeedDiffGroup bool) sarama.ConsumerGroup {
	onceConsumer.Do(func() {
		var kafkaUrl string
		kafkaUrl, err := panda_nacos.GetNacosClient().GetConfig(ctx, &panda_nacos.ConfigParam{
			DataId: panda_nacos.ServerName + ".kafka.config",
		})
		if err != nil {
			log.Warn(ctx, "[GetKafkaConsumer] get current service kafka config fail", zap.Error(err))
			kafkaUrl, err = panda_nacos.GetNacosClient().GetConfig(ctx, &panda_nacos.ConfigParam{
				DataId: "common.kafka.config",
				Group:  "DEFAULT_GROUP",
			})
			if err != nil {
				log.Fatal(ctx, "[GetKafkaConsumer] get kafka config all fail", zap.Error(err))
			}
			if kafkaUrl == "" {
				log.Fatal(ctx, "[GetKafkaConsumer] get common kafka config empty", zap.Error(err))
			}
		}
		if kafkaUrl == "" {
			log.Fatal(ctx, "[GetKafkaConsumer] get service kafka config empty", zap.String("service_name", panda_nacos.ServerName), zap.Error(err))
		}
		var groupName string
		if isNeedDiffGroup {
			groupName = panda_nacos.PodName + "_consumer_group"
		} else {
			groupName = panda_nacos.ServerName + "_consumer_group"
		}
		config := sarama.NewConfig()
		config.Consumer.Return.Errors = true
		consumerGroup, err = sarama.NewConsumerGroup([]string{kafkaUrl}, groupName, config)
		if err != nil {
			log.Fatal(ctx, "[GetKafkaConsumer]  Error creating consumer group", zap.Error(err))
			return
		}
	})
	return consumerGroup
}

// GetTopicMessageChan isNeedDiffGroup:是否创建以pod为名字的消费者组,还是以deployment名字创建
func GetTopicMessageChan(ctx context.Context, topic string, isNeedDiffGroup bool) (<-chan *sarama.ConsumerMessage, error) {
	chanVal, isGet := topic2MessageChanMap.Load(topic)
	if isGet {
		if cc, ok := chanVal.(chan *sarama.ConsumerMessage); ok {
			return cc, nil
		}
	}
	myTopicChan := make(chan *sarama.ConsumerMessage, 10)
	consumerHandler := &ConsumerHandler{
		myTopicChan,
	}
	go func() {
		err := GetKafkaConsumer(ctx, isNeedDiffGroup).Consume(ctx, []string{topic}, consumerHandler)
		if err != nil {
			log.Error(ctx, "[GetTopicMessageChan] Consume fail", zap.String("topic", topic), zap.Error(err))
		}
	}()

	topic2MessageChanMap.Store(topic, myTopicChan)
	return myTopicChan, nil
}

// ConsumerHandler 实现了 sarama.ConsumerGroupHandler 接口
type ConsumerHandler struct {
	messages chan *sarama.ConsumerMessage
}

func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		// 将消息发送到通道
		h.messages <- msg
		session.MarkMessage(msg, "")
	}
	return nil
}

//
//func main() {
//	brokerAddress := "kafka.common.svc.cluster.local:9092" // 指定 Kafka 集群中任一 Broker 的地址和端口
//	topic := "your_topic"                                  // 指定要订阅的主题
//	// 配置 SASL/SCRAM 认证参数
//	config := sarama.NewConfig()
//	//config.Net.SASL.Enable = true
//	//config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
//	//config.Net.SASL.User = "user1"
//	//config.Net.SASL.Password = "ozSUKGWtHZ"
//	//config.Net.SASL.Handshake = true
//	// 创建 Kafka 消费者
//	ctx := context.TODO()
//	consumer, err := sarama.NewConsumer([]string{brokerAddress}, config)
//	if err != nil {
//		log.Fatal(ctx, "Error creating consumer:", zap.Error(err))
//	}
//	if err != nil {
//		fmt.Printf("Error creating consumer group: %v\n", err)
//		return
//	}
//
//	// 创建一个分区消费者
//	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
//	if err != nil {
//		log.Fatal(ctx, "Error creating partition consumer:", zap.Error(err))
//	}
//	// 处理消息
//	var wg sync.WaitGroup
//	wg.Add(1)
//
//	go func() {
//		defer wg.Done()
//		for {
//			select {
//			case msg := <-partitionConsumer.Messages():
//				fmt.Printf("Received message: Topic=%s, Partition=%d, Offset=%d, Key=%s, Value=%s\n",
//					msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
//			case err := <-partitionConsumer.Errors():
//				fmt.Println("Error:", err)
//			}
//		}
//	}()
//
//	// 等待信号，例如 Ctrl+C，然后关闭消费者
//	sigterm := make(chan os.Signal, 1)
//	signal.Notify(sigterm, os.Interrupt)
//	select {
//	case <-sigterm:
//		fmt.Println("Interrupted. Closing consumer.")
//		partitionConsumer.Close()
//	}
//
//	// 等待消息处理协程结束
//	wg.Wait()
//
//	// 关闭消费者
//	if err := consumer.Close(); err != nil {
//		log.Fatal(ctx, "Error closing consumer:", zap.Error(err))
//	}
//}

//
//package main
//
//import (
//	"context"
//	"go.uber.org/zap"
//	"panda-go/common/log"
//	"panda-go/component/k8s"
//	"time"
//
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//)
//
//func main() {
//	time.Sleep(10 * time.Second)
//	ctx := context.TODO()
//	clientset := k8s.GetK8sClient()
//	if clientset != nil {
//		for {
//			services, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
//			if err != nil {
//				log.Error(ctx, "GetNewServiceEntity call to k8s api fail", zap.Error(err))
//			}
//			if services != nil {
//				for _, service := range services.Items {
//					log.Info(ctx, "get service name:", zap.String("service_name", service.Name))
//				}
//			}
//			//pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
//			//if err != nil {
//			//	fmt.Println(err)
//			//} else {
//			//	log.Infof(ctx, "There are %d pods in the cluster\n", len(pods.Items))
//			//	for _, pod := range pods.Items {
//			//		log.Info(ctx, "get pod name:", zap.String("pod_name", pod.Name))
//			//	}
//			//}
//			time.Sleep(100000 * time.Second)
//		}
//	}
//}
