package snowflake

import (
	"errors"
	"sync"
	"time"
)

type Worker struct {
	dataCenterId uint64 //  数据中心
	machineId    uint64 //  机器标识
	sequence     uint64 //  序列号
	lastStamp    uint64 // 上一次时间戳
	lock         sync.Mutex
}

func NewWorker(dataCenterId, machineId uint64) (*Worker, error) {
	cmu.Lock()
	defer cmu.Unlock()
	if dataCenterId > MAX_DATA_CENTER_NUM || dataCenterId < 0 {
		return nil, errors.New("dataCenterId can't be greater than MAX_DATA_CENTER_NUM or less than 0")
	}
	if machineId > MAX_MACHINE_NUM || machineId < 0 {
		return nil, errors.New("machineId can't be greater than MAX_MACHINE_NUM or less than 0")
	}
	return &Worker{
		dataCenterId: dataCenterId,
		machineId:    machineId,
		sequence:     0,
		lastStamp:    -1,
	}, nil
}

func (w *Worker) nextMills() uint64 {
	currStamp := currentTimeMills()
	for currStamp < w.lastStamp {
		currStamp = currentTimeMills()
	}
	return currStamp
}

func (w *Worker) NextBatch(count int) ([]uint64, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	if count <= 1 {
		if v, err := w.nextId(); err == nil {
			return []uint64{v}, nil
		} else {
			return nil, err
		}
	}
	ret := make([]uint64, count)
	for i := 0; i < count; i++ {
		if v, err := w.nextId(); err == nil {
			ret[i] = v
		} else {
			return nil, err
		}
	}
	return ret, nil
}

func (w *Worker) NextId() (uint64, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.nextId()
}

func (w *Worker) nextId() (uint64, error) {
	currStamp := currentTimeMills()
	if currStamp < w.lastStamp {
		return 0, errors.New("Clock moved backwards.  Refusing to generate id")
	}
	if currStamp == w.lastStamp {
		// 相同毫秒内，序列号自增
		w.sequence = (w.sequence + 1) & MAX_SEQUENCE
		// 同一毫秒的序列数已经达到最大
		if w.sequence == 0 {
			currStamp = w.nextMills()
		}
	} else {
		// 不同毫秒内，序列号置为0
		w.sequence = 0
	}

	w.lastStamp = currStamp
	// 时间戳部分 | 数据中心部分 | 机器标识部分 | 序列号部分
	return (currStamp-START_STAMP)<<TIMESTAMP_LEFT | w.dataCenterId<<DATA_CENTER_LEFT | w.machineId<<MACHINE_LEFT | w.sequence, nil
}

func currentTimeMills() uint64 {
	return uint64(time.Now().UnixNano() / 1000000)
}
