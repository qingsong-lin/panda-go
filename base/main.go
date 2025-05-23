package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"os"
	"panda-go/common/grpc-wrapper/grpc_server_wrapper/holms_wrapper"
	"panda-go/common/log"
	"panda-go/common/utils"
	"panda-go/component/k8s"
	panda_nacos "panda-go/component/nacos/base"
	"reflect"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"
)

type MyStruct1 struct {
	name string
	age  int
}
type MyStruct2 struct {
	address string
}

func (s *MyStruct1) Name(string2 string) {

}

type name interface {
	Name()
}

//func changeInterface(t *name) {
//	if entity, ok := *t.(*MyStruct1); ok {
//		entity.name = "16346546"
//		println(entity.age)
//	}
//}

func changeTuples(in *[3]MyStruct1) {
	in[0] = MyStruct1{
		name: "99999999",
	}
}

type LgxMu struct {
	Name string
	mu   sync.Mutex
}

func func1(a interface{}) {
	// 利用反射获取结构体指针的元素类型
	val := reflect.ValueOf(a).Elem()

	// 获取结构体类型
	structType := val.Type()

	// 遍历结构体的字段
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := structType.Field(i)

		// 判断字段是否可导出
		if fieldType.PkgPath == "" {
			// 判断字段是否为字符串类型
			if fieldType.Type.Kind() == reflect.String {
				// 设置字段的新值
				field.SetString("hahahaha")
			}
		}
	}
}

type LgxMu2 struct {
	name string
}

//type ConfigParam struct {
//	DataId           string                                      `param:"dataId"`  //required
//	Group            string                                      `param:"group"`   //required
//	Content          string                                      `param:"content"` //required
//	Tag              string                                      `param:"tag"`
//	AppName          string                                      `param:"appName"`
//	BetaIps          string                                      `param:"betaIps"`
//	CasMd5           string                                      `param:"casMd5"`
//	Type             string                                      `param:"type"`
//	SrcUser          string                                      `param:"srcUser"`
//	EncryptedDataKey string                                      `param:"encryptedDataKey"`
//	KmsKeyId         string                                      `param:"kmsKeyId"`
//	OnChange         func(namespace, group, dataId, data string) `param:"onChange"`
//}

type Mutex struct {
}

func (m *Mutex) Name() string {
	return "mem"
}

func (m *Mutex) Run() {
	mutex := &sync.Mutex{}
	// 这里模拟了死锁的情况
	mutex.Lock()
	go func() {
		time.Sleep(time.Second)
		mutex.Unlock()
	}()
	mutex.Lock()

}

type TokenBucket struct {
	rate       float64 // 每秒产生的令牌数
	capacity   float64 // 令牌桶容量
	tokens     float64 // 当前令牌数
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucket(rate, capacity float64) *TokenBucket {
	return &TokenBucket{
		rate:     rate,
		capacity: capacity,
		tokens:   capacity,
	}
}

func (tb *TokenBucket) Allow(needToken float64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	tmpToken := tb.tokens + tb.rate*now.Sub(tb.lastRefill).Seconds() - needToken

	if tmpToken > tb.capacity {
		tmpToken = tb.capacity
	}

	if tmpToken >= 0 {
		tb.tokens = tmpToken
		tb.lastRefill = now
		return true
	}
	return false
}

type LeakyBucket struct {
	rate            float64 // 漏桶速率，每秒处理的请求数
	bucketSize      float64 // 漏桶容量，即最大存储请求数
	storage2Process float64 // 漏桶中需要处理的令牌数
	lastLeakTime    time.Time
	mu              sync.Mutex
}

func NewLeakyBucket(rate, bucketSize float64) *LeakyBucket {
	return &LeakyBucket{
		rate:            rate,
		bucketSize:      bucketSize,
		storage2Process: 0,
		lastLeakTime:    time.Now(),
	}
}

// Allow 漏铜算法
func (lb *LeakyBucket) Allow(requiredTokens float64) bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	// 计算时间间隔，漏掉的令牌数量
	elapsed := time.Since(lb.lastLeakTime).Seconds()
	leakedTokens := elapsed * lb.rate

	tmpStorage2Process := lb.storage2Process + leakedTokens + requiredTokens
	// 计算本次请求需要的令牌数量
	if tmpStorage2Process > lb.bucketSize {
		fmt.Printf("%f ", lb.storage2Process)
		return false
	}

	if tmpStorage2Process <= 0 {
		lb.storage2Process = 0
	} else {
		lb.storage2Process = tmpStorage2Process
	}
	fmt.Printf("%f ", lb.storage2Process)
	lb.lastLeakTime = time.Now()

	return true
}

// User 模型
type User struct {
	gorm.Model
	Name    string
	Email   string
	Profile Profile // 一对一关联
	Orders  []Order // 一对多关联
	Roles   []Role  `gorm:"many2many:user_roles;"` // 多对多关联
}

// Profile 模型，一对一关联
type Profile struct {
	gorm.Model
	UserID uint
	Bio    string
}

// Order 模型，一对多关联
type Order struct {
	gorm.Model
	UserID    uint
	Product   string
	OrderDate string
}

// Role 模型，多对多关联
type Role struct {
	gorm.Model
	Name  string
	Users []User `gorm:"many2many:user_roles;"` // 多对多关联
}

func getCurrentGoRoutineId() int64 {
	var buf [64]byte
	// 获取栈信息
	n := runtime.Stack(buf[:], false)
	// 抽取id
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine"))[0]
	// 转为64位整数
	id, _ := strconv.Atoi(idField)
	return int64(id)
}
func testUsersIndexFunc(obj interface{}) ([]string, error) {
	pod := obj.(*v1.Pod)
	usersString := pod.Annotations["users"]

	return strings.Split(usersString, ","), nil
}
func TestMultiIndexKeys(t *testing.T) {

	//jobInformer.AddIndexers(cache.Indexers{"byUser": UsersIndexFunc})
	//jobInformer := cache.NewSharedIndexInformer(
	//	&cache.ListWatch{
	//		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
	//			options.LabelSelector = labelSelector
	//			return kclient.BatchV1().Jobs(namespace).List(ctx, options)
	//		},
	//		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
	//			options.LabelSelector = labelSelector
	//			return kclient.BatchV1().Jobs(namespace).Watch(ctx, options)
	//		},
	//	},
	//	&batchv1.Job{},
	//	resyncPeriod,
	//	cache.Indexers{},
	//)
	// Create informer for watching Namespaces
	//podInformer := cache.NewSharedIndexInformer(
	//	&cache.ListWatch{
	//		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
	//			options.LabelSelector = labelSelector
	//			return kclient.CoreV1().Pods(namespace).List(ctx, options)
	//		},
	//		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
	//			options.LabelSelector = labelSelector
	//			return kclient.CoreV1().Pods(namespace).Watch(ctx, options)
	//		},
	//	},
	//	&corev1.Pod{},
	//	resyncPeriod,
	//	cache.Indexers{},
	//)

	index := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{"byUser": testUsersIndexFunc})
	sharedInformers := informers.NewSharedInformerFactoryWithOptions(k8s.GetK8sClient(), time.Second*30, informers.WithTweakListOptions(func(opt *metav1.ListOptions) {
		opt.LabelSelector = "create-by=panda"
	}))
	jobInformer := sharedInformers.Batch().V1().Jobs().Informer()
	jobInformer.AddIndexers(cache.Indexers{"byUser": testUsersIndexFunc})

	pod1 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "one", Annotations: map[string]string{"users": "ernie,bert"}}}
	pod2 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "two", Annotations: map[string]string{"users": "bert,oscar"}}}
	pod3 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "tre", Annotations: map[string]string{"users": "ernie,elmo"}}}
	jobInformer.GetIndexer().Add(pod1)
	indexResults, err := jobInformer.GetIndexer().ByIndex("byUser", "ernie")
	println(len(indexResults))
	index.Add(pod1)
	index.Add(pod2)
	index.Add(pod3)

	expected := map[string]sets.String{}
	expected["ernie"] = sets.NewString("one", "tre")
	expected["bert"] = sets.NewString("one", "two")
	expected["elmo"] = sets.NewString("tre")
	expected["oscar"] = sets.NewString("two")
	expected["elmo1"] = sets.NewString()
	{
		for k, v := range expected {
			found := sets.String{}
			indexResults, err := index.ByIndex("byUser", k)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}
			for _, item := range indexResults {
				found.Insert(item.(*v1.Pod).Name)
			}
			if !found.Equal(v) {
				t.Errorf("missing items, index %s, expected %v but found %v", k, v.List(), found.List())
			}
		}
	}

	index.Delete(pod3)
	erniePods, err := index.ByIndex("byUser", "ernie")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(erniePods) != 1 {
		t.Errorf("Expected 1 pods but got %v", len(erniePods))
	}
	for _, erniePod := range erniePods {
		if erniePod.(*v1.Pod).Name != "one" {
			t.Errorf("Expected only 'one' but got %s", erniePod.(*v1.Pod).Name)
		}
	}

	elmoPods, err := index.ByIndex("byUser", "elmo")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(elmoPods) != 0 {
		t.Errorf("Expected 0 pods but got %v", len(elmoPods))
	}

	copyOfPod2 := pod2.DeepCopy()
	copyOfPod2.Annotations["users"] = "oscar"
	index.Update(copyOfPod2)
	bertPods, err := index.ByIndex("byUser", "bert")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(bertPods) != 1 {
		t.Errorf("Expected 1 pods but got %v", len(bertPods))
	}
	for _, bertPod := range bertPods {
		if bertPod.(*v1.Pod).Name != "one" {
			t.Errorf("Expected only 'one' but got %s", bertPod.(*v1.Pod).Name)
		}
	}
}

func deferFunc() {
	defer println("first")
	defer println("hhahah")
}

func init() {

}

func init() {

}

type duiqi struct {
	a int8
	b unsafe.Pointer
}

type LgxMu3 struct {
	Name []string
}

func testSSTTU() {
	strstr := "{\"name\":[\"jhhahah\"],\"lgx\":\"nponmomno\"}"
	jsonLgx := LgxMu3{}
	json.Unmarshal([]byte(strstr), &jsonLgx)
	println(jsonLgx.Name)
}

func main() {
	testSSTTU()
	tma := duiqi{}
	println(unsafe.Sizeof(tma))

	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			fmt.Println("i: ", i)
			wg.Done()
		}(i)
	}
	wg.Wait()
	e := []int32{1, 2, 3}
	e = append(e, 4, 5, 6, 7)
	fmt.Println("cap of e after:", cap(e))
	//pool := grpool.New(100)
	//
	////添加1千个任务
	//for i := 0; i < 1000; i++ {
	//	_ = pool.Add(job)
	//}
	//TestMultiIndexKeys(nil)
	var cc chan string
	cc = make(chan string)
	go func() {
		<-cc
	}()
	cc <- "st"
	tmpst := struct {
	}{}
	println(unsafe.Sizeof(tmpst))
	deferFunc()
	reg := regexp.MustCompile(`\(\d+(%(\d*(\.)?\d*)?[a-z])\)`)
	ttm := reg.ReplaceAllString("lgx(1%s)hahh(2%s)", "${1}")
	println(len(ttm))
	re := regexp.MustCompile("([0-9]+)")
	m := re.FindAllString("lsf1235sdf2s15", -1)
	println(m[0])
	var intValue *int
	intValue = new(int)
	println(intValue)
	router := gin.Default()
	router.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, World!"})
	})
	router.GET("/hlo", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, World!"})
	})
	router.GET("/h122lo", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, World!"})
	})
	router.Run(":8089")
	go func() {
		getCurrentGoRoutineId()
		time.Sleep(time.Hour)
	}()
	time.Sleep(time.Hour)
	dsn := "root:panda@tcp(192.168.50.20:32144)/panda?charset=utf8mb4&parseTime=True&loc=Local"

	// 连接到 MySQL 数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(context.TODO(), "Set ", zap.Error(err))
	}
	err = db.AutoMigrate(&User{}, &Profile{}, &Role{}, &Order{})
	if err != nil {
		log.Fatal(context.TODO(), "Set ", zap.Error(err))
	}
	db.Create(&User{Name: "John", Email: "john@example.com", Profile: Profile{Bio: "Gopher"}})

	// 示例：创建一对多关联
	user := User{Name: "Alice", Email: "alice@example.com"}
	order1 := Order{Product: "Product A", OrderDate: "2022-01-01"}
	order2 := Order{Product: "Product B", OrderDate: "2022-01-02"}
	user.Orders = []Order{order1, order2}
	db.Create(&user)

	// 示例：创建多对多关联
	user1 := User{Name: "Bob", Email: "bob@example.com"}
	user2 := User{Name: "Charlie", Email: "charlie@example.com"}
	role1 := Role{Name: "Admin"}
	role2 := Role{Name: "Editor"}
	user1.Roles = []Role{role1}
	user2.Roles = []Role{role2}
	db.Create(&user1)
	db.Create(&user2)

	// 查询一对一关联
	//var myUser []User
	var userlgx User
	//var tmplgx []Profile
	db.First(&userlgx, "name = ?", "John")
	db.Model(userlgx).Association("Profile").Find(&userlgx.Profile)

	// 查询一对多关联
	var userWithOrders User
	db.Preload("Orders").First(&userWithOrders, "name = ?", "Alice")

	// 查询多对多关联
	var userWithRoles User
	db.Preload("Roles").First(&userWithRoles, "name = ?", "Bob")

	// 自动迁移表结构

	//leakyBucket := NewLeakyBucket(10, 50)
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    []string{"lgxcloud.com:30530"}, // Redis 服务器地址
		Password: "64519385",                     // Redis 服务器密码
	})

	// 测试字符串类型
	err = client.Set(context.TODO(), "golang-test", "John Doe", 0).Err()
	if err != nil {
		log.Fatal(context.TODO(), "Set ", zap.Error(err))
	}

	needTokens := []float64{10, 20, 30, 10, 40, 5, 2, 1}
	tokenBucket := NewTokenBucket(10, 50)
	// 模拟请求
	for i := 1; i <= 30; i++ {
		tokens := needTokens[rand.Intn(7)]
		if tokenBucket.Allow(tokens) {
			fmt.Printf("Request %d allowed at %s need token %f\n", i, time.Now().Format("2006-01-02 15:04:05"), tokens)
		} else {
			fmt.Printf("Request %d blocked at %s need token %f\n", i, time.Now().Format("2006-01-02 15:04:05"), tokens)
		}

		time.Sleep(200 * time.Millisecond) // 模拟请求间隔
	}
	time.Sleep(time.Hour)
	//var wg sync.WaitGroup
	go func() {
		tmp := &Mutex{}
		tmp.Run()
	}()
	holms_wrapper.StartHolmes()
	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)
	for _, pprofType := range []string{"goroutine"} {
		file, err := zipWriter.Create(pprofType)
		if err != nil {
			log.Error(context.TODO(), "[DownloadPprofFile] zipWriter.Create fail", zap.Error(err))
		}
		var buff bytes.Buffer
		_ = pprof.Lookup(pprofType).WriteTo(&buff, 0)
		file.Write(buff.Bytes())
	}
	zipWriter.Close()
	// 设置响应头，告诉浏览器这是一个ZIP文件下载

	content := zipBuffer.Bytes()
	println(content)
	time.Sleep(5 * time.Second)
	// 开启对锁调用的跟踪，不开启的话抓取不到（下同）
	for _, look := range []string{"goroutine", "heap", "threadcreate", "block"} {
		file, err := os.Create(look)
		if err != nil {
			println(err)
		}
		_ = pprof.Lookup(look).WriteTo(file, 0)
		file.Close()
	}
	for i := 1; i < 5; i++ {
		go func() {
			time.Sleep(time.Hour)
		}()
	}

	//go func() {
	//	r := gin.Default()
	//	// 定义路由处理程序
	//	r.GET("/test", func(c *gin.Context) {
	//		c.JSON(http.StatusOK, gin.H{
	//			"message": "Hello, World!",
	//		})
	//	})
	//	r.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))
	//	// 启动 HTTP 服务器
	//	if err := r.Run(":8080"); err != nil {
	//		panic(err)
	//	}
	//
	//}()
	//_ = http.ListenAndServe(":8080", nil)
	//go func() {
	//	m := &Mutex{}
	//	m.Run()
	//}()
	//// 创建一个等待组，用于等待所有 goroutine 完成
	//for i := 0; i < 5; i++ {
	//	wg.Add(1)
	//	go func(id int) {
	//		defer wg.Done()
	//
	//		// 在这里模拟一些耗时的操作，导致阻塞
	//		time.Sleep(time.Second * 100000)
	//
	//		fmt.Printf("Goroutine %d completed.\n", id)
	//	}(i)
	//}
	//
	//// 等待所有 goroutine 完成
	//wg.Wait()

	//time.Sleep(time.Hour)
	//zipBuffer := new(bytes.Buffer)
	//zipWriter := zip.NewWriter(zipBuffer)
	//defer zipWriter.Close()
	//for _, pprofType := range []string{"goroutine"} {
	//	file, err := zipWriter.Create(pprofType)
	//	if err != nil {
	//		log.Error(context.TODO(), "[DownloadPprofFile] zipWriter.Create fail", zap.Error(err))
	//	}
	//	err = pprof.Lookup(pprofType).WriteTo(file, 0)
	//	println(err)
	//}
	//file, err := os.Create("lgxpprof.zip")
	//if err != nil {
	//	println(err)
	//}
	//defer file.Close()
	//
	//_, err = zipBuffer.WriteTo(file)
	//println(err)

	inParam := &vo.ConfigParam{
		DataId:           "sdjflks",
		Group:            "",
		Content:          "",
		Tag:              "",
		AppName:          "",
		BetaIps:          "",
		CasMd5:           "",
		Type:             "",
		SrcUser:          "",
		EncryptedDataKey: "",
		KmsKeyId:         "",
		UsageType:        "",
		OnChange: func(namespace, group, dataId, data string) {

		},
	}
	log.Info(context.TODO(), "Nacos: GetConfig success", zap.Any("inParam", inParam))
	config, err := panda_nacos.GetNacosClient().PublishConfig(context.TODO(), &panda_nacos.ConfigParam{
		DataId:  "lgtxxx",
		Group:   "DEV",
		Content: "hahahaha",
		AppName: "lgxpandaapp",
	}, "")
	if err != nil {
		return
	}
	println(config)
	//tmpconfig := "{\"ClientConfig\":{\"NamespaceId\":\"production\",\"Username\":\"nacos\",\"Password\":\"panda\"}, \"ServerConfigs\": [{\"IpAddr\":\"127.0.0.1\", \"Port\": 8848}]}"
	//log.Info(context.TODO(), "tmpconfig", zap.String("tmpconfig", tmpconfig))
	//nacosParam := vo.NacosClientParam{}
	//err := json.Unmarshal([]byte(tmpconfig), &nacosParam)
	//if err != nil {
	//
	//}
	//sc := []constant.ServerConfig{
	//	*constant.NewServerConfig("127.0.0.1", 8848, constant.WithContextPath("/nacos")),
	//	//*constant.NewServerConfig("192.168.50.20", 32735, constant.WithContextPath("/nacos")),
	//	//*constant.NewServerConfig("192.168.50.20", 30769, constant.WithContextPath("/nacos")),
	//	//*constant.NewServerConfig("192.168.50.20", 31007, constant.WithContextPath("/nacos")),
	//}
	////设置namespace的id    日志目录
	//cc := *constant.NewClientConfig(
	//	constant.WithNamespaceId("production"),
	//	constant.WithTimeoutMs(5000),
	//	constant.WithNotLoadCacheAtStart(true),
	//	constant.WithLogDir("tmp/nacos/log"),
	//	constant.WithCacheDir("tmp/nacos/cache"),
	//	constant.WithLogLevel("debug"),
	//)
	////建立连接
	//cc.Username = "nacos"
	//cc.Password = "panda"
	//client, err := clients.NewConfigClient(nacosParam) //vo.NacosClientParam{
	//	ClientConfig:  &cc,
	//	ServerConfigs: sc,
	//},

	//if err != nil {
	//	fmt.Printf("PublishConfig err:%+v \n", err)
	//}
	////获取配置集
	//content, err := client.PublishConfig(vo.ConfigParam{
	//	DataId:  "lgxtmp",
	//	Group:   "DEFAULT_GROUP",
	//	Content: "hahahahaha",
	//})
	//fmt.Println(content)
	//这里是自己实例化的struct

	//fmt.Println(content)

	//tmpconfig := "{\"ClientConfig\":{\"NamespaceId\":\"production\",\"Username\":\"nacos\",\"Password\":\"panda\"}, \"ServerConfigs\": [{\"IpAddr\":\"nacos-cs.common\", \"Port\": 8848}]}"
	//log.Info(context.TODO(), "tmpconfig", zap.String("tmpconfig", tmpconfig))
	//nacosParam := vo.NacosClientParam{}
	//err = json.Unmarshal([]byte(tmpconfig), &nacosParam)
	//if err != nil {
	//
	//}
	myMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	// 将 map[string]string 转换为 JSON
	jsonData, err := json.Marshal(myMap)
	if err != nil {
		fmt.Println("JSON encoding error:", err)
		return
	}
	//configParam := &ConfigParam{
	//	//DataId: "213",
	//}
	//strfunc, err := json.Marshal(configParam)
	//log.Info(context.TODO(), "err", zap.Any("param", strfunc))
	// 输出 JSON 字符串
	fmt.Println(string(jsonData))

	// 将 JSON 字符串解码为 sync.Map
	var syncMap sync.Map
	syncMap.Store("12", "hahaha")
	outsyncmap, _ := json.Marshal(syncMap)
	println(outsyncmap)
	err = json.Unmarshal(jsonData, &syncMap)
	if err != nil {
		fmt.Println("JSON decoding error:", err)
		return
	}
	tmpssss, ok := syncMap.Load("key1")
	if val, ok2 := tmpssss.(string); ok && ok2 {
		println(val)
	}
	//a := LgxMu{Name: "Original Value"}
	//fmt.Printf("Before: %+v\n", a)
	//
	//// 将结构体变量传递给 func1 函数
	//func1(a)
	//
	////// 打印更新后的值
	//fmt.Printf("After: %+v\n", a)
	strstr := "{\"name\":\"jhhahah\"}"
	jsonLgx := LgxMu2{}
	json.Unmarshal([]byte(strstr), &jsonLgx)
	println(jsonLgx.name)

	hahah := LgxMu{
		Name: "lgx",
		mu:   sync.Mutex{},
	}
	hahah.mu.Lock()
	hahah.mu.TryLock()
	hahah.mu.Unlock()
	sliceMap := make(map[string][]string)
	sliceMap["lgx"] = []string{"1", "2"}
	if ss, ok := sliceMap["lgx"]; ok {
		ss[1] = "hahahaha"
		ss = append(ss, "152412")
	}
	hahah.mu.Lock()
	valDirectType := reflect.Indirect(reflect.ValueOf(hahah)).Type()
	newVal := reflect.New(valDirectType).Interface()
	_ = copier.CopyWithOption(newVal, hahah, copier.Option{DeepCopy: true})
	copyHahainter := utils.DeepCopyV2(&hahah)
	//copyHahainter := utils.DeepCopyV2(hahah)
	//copyHahainter
	if copyHaha, ok := copyHahainter.(LgxMu); ok {
		copyHaha.mu.Lock()
		println(copyHaha.Name)
		hahah.Name = "wanger"
		println(copyHaha.Name)
	}

	tmpFunc := (*MyStruct1).Name
	tmpFunc(nil, "")
	log.Debug(context.Background(), "Error connecting to Server B:")
	//log.SetLevel(zapcore.DebugLevel)
	log.Debug(context.Background(), "Error connecting to Server B:")
	//logrus.SetLevel(logrus.DebugLevel)
	//logrus.Debugln("hahahahah")
	//logrus.SetLevel(logrus.InfoLevel)
	//logrus.Debugln("222222")
	tmpSlice := []int{1, 2, 4, 3}
	tt := tmpSlice[:200]
	println(tt)
	first := MyStruct1{
		name: "41564654",
	}
	//changeInterface(&first)
	var tmp *[3]MyStruct1
	//tmp[0] = first
	changeTuples(tmp)
	println(first.name)
	//ch := make(chan *pubsub.Message, 2)
	//go func() {
	//	for {
	//		select {
	//		case msg := <-ch:
	//			println(msg.Content)
	//		default:
	//
	//		}
	//	}
	//
	//}()
	//go func() {
	//	for {
	//		select {
	//		case ch <- &pubsub.Message{
	//			Content: "test",
	//		}:
	//			println("success")
	//		default:
	//
	//		}
	//	}
	//}()
	//time.Sleep(1 * time.Hour)
	m1 := MyStruct1{
		name: "lgx",
		age:  18,
	}
	in := reflect.ValueOf(m1)
	inType := in.Type()
	for i := 0; i < in.NumField(); i++ {
		field := in.Field(i)
		if !field.CanSet() {
			if !field.CanInterface() {
				// 如果字段不可导出，尝试通过字段名获取
				field = in.FieldByName(inType.Field(i).Name)
			}
		}

	}
	str, _ := json.Marshal(m1)
	println(str)
	m2 := testDeepCopy(m1)
	if m2Val, ok := m2.(*MyStruct1); ok {
		m2Val2 := *m2Val
		m2Val2.name = "hahah"
	}
	//println(m1)
	GetStructsInPackage("main")
}

func testDeepCopy(in interface{}) (re interface{}) {
	//re, _ = utils.DeepCopy(in)
	re = utils.DeepCopyV2(in)
	return
}

func GetStructsInPackage(pkgName string) []reflect.Type {
	pkg := reflect.ValueOf(pkgName)
	pkgType := pkg.Type()
	var structs []reflect.Type
	for i := 0; i < pkg.NumField(); i++ {
		field := pkg.Field(i)
		fieldType := pkgType.Field(i)
		if fieldType.Type.Kind() == reflect.Struct {
			structs = append(structs, field.Type())
		}
	}
	fmt.Println("Structs in package:", pkgName, structs)
	return structs
}
