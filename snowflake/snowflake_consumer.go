package snowflake

import (
	"errors"
	jsoniter "github.com/json-iterator/go"
	"github.com/pojozhang/sugar"
	"github.com/ywengineer/snowflake-golang/math"
	"go.uber.org/zap"
	nh "net/http"
	"net/url"
	"sync"
	"time"
)

var http = sugar.New(func() sugar.Transporter {
	return &nh.Client{
		Timeout: 5 * time.Second,
	}
})

type SnowflakeConsumer struct {
	destination   string
	headers       map[string]interface{}
	paddingFactor int
	queue         *BlockingQueue
	lock          sync.Mutex
	log           *zap.Logger
	retry         int
}

func NewConsumer(dest string, headers map[string]interface{}, queueSize, paddingFactor, retry int, log *zap.Logger) (*SnowflakeConsumer, error) {
	if len(dest) == 0 {
		return nil, errors.New("snowflake server destination is not present")
	}
	if _, err := url.ParseRequestURI(dest); err != nil {
		return nil, err
	}
	queueSize = math.MaxInt(4096, queueSize)
	factor := float64(queueSize) * (float64(100-math.MinInt(100, paddingFactor)) / 100)
	consumer := &SnowflakeConsumer{
		destination:   dest,
		headers:       headers,
		paddingFactor: int(factor),
		retry:         math.MaxInt(retry, 2),
		queue:         NewQueue(queueSize),
		lock:          sync.Mutex{},
		log:           log,
	}
	consumer.retrieve()
	return consumer, nil
}

func (c *SnowflakeConsumer) _consume() (int64, error) {
	//
	if c.queue.Len() == 0 {
		//
		c.retrieve()
		//
		return c.queue.Take()
	}
	if value, err := c.queue.Take(); err != nil {
		return value, err
	} else {
		c.retrieve()
		return value, nil
	}
}

func (c *SnowflakeConsumer) Consume() (int64, error) {
	for count := 0; count < c.retry; count++ {
		if v, err := c._consume(); err == nil {
			return v, nil
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}
	//
	return 0, nil
}

func (c *SnowflakeConsumer) isNotSufficient() bool {
	// 剩余空间
	remain := c.queue.RemainCapacity()
	//
	if remain >= c.paddingFactor {
		c.log.Info("fill id event", zap.Int("remain", remain), zap.Int("paddingFactor", c.paddingFactor))
	}
	// 如果队列剩余容量达到补充比例
	return remain >= c.paddingFactor
}

func (c *SnowflakeConsumer) retrieve() {
	go func() {
		c.lock.Lock()
		defer c.lock.Unlock()
		if c.queue.Len() == 0 || c.isNotSufficient() {
			// 计算数量，
			size := c.queue.RemainCapacity() //math.MaxInt(c.queue.Capacity()-c.paddingFactor, c.queue.RemainCapacity())
			//
			if b, res, err := http.Get(c.destination,
				sugar.Q{"size": size},
				sugar.H(c.headers)).ReadBytes(); err != nil {
				c.log.Error("invoke snowflake server dest error", zap.Error(err))
			} else {
				if res.StatusCode/100 == 2 {
					ret := &SnowflakeApiResult{}
					if err := jsoniter.Unmarshal(b, &ret); err != nil {
						c.log.Error("parse snowflake api result error", zap.Error(err))
					} else if len(ret.Data) > 0 {
						c.queue.PushAll(ret.Data)
					}
				} else {
					c.log.Error("invoke snowflake remote destination error", zap.Int("status", res.StatusCode), zap.String("body", string(b)))
				}
			}
		}
	}()
}
