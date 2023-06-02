package snowflake

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	epoch          int64 = 1672502400000               // 2023-01-01 00:00:00.000
	workerIdBits   uint8 = 9                           // 机器ID的位数
	sequenceBits   uint8 = 14                          // 序列号的位数
	maxWorkerId    int64 = -1 ^ (-1 << workerIdBits)   // 机器ID的最大值
	maxSequence    int64 = -1 ^ (-1 << sequenceBits)   // 序列号的最大值
	workerIdShift  uint8 = sequenceBits                // 机器ID左移位数
	timestampShift uint8 = sequenceBits + workerIdBits // 时间戳左移位数
)

type SnowFlake struct {
	workerId      int64
	sequence      int64
	lastTimestamp int64
	mutex         sync.Mutex
}

func NewSnowFlake(workerId int64) (*SnowFlake, error) {
	if workerId < 0 || workerId > maxWorkerId {
		return nil, errors.New(fmt.Sprintf("workerId必须在0-%d之间", maxWorkerId))
	}
	return &SnowFlake{
		workerId: workerId,
	}, nil
}

// NextId 获取下一个ID
func (snowFlake *SnowFlake) NextId() int64 {
	snowFlake.mutex.Lock()
	defer snowFlake.mutex.Unlock()

	timestamp := time.Now().UnixNano() / 1000000
	// 发生了时钟回拨，等待时钟追上
	if timestamp < snowFlake.lastTimestamp {
		timestamp = snowFlake.wait()
	}

	if timestamp == snowFlake.lastTimestamp {
		snowFlake.sequence = (snowFlake.sequence + 1) & maxSequence
		// 序列号溢出，等待下一毫秒
		if snowFlake.sequence == 0 {
			timestamp = snowFlake.waitNext()
		}
	} else {
		snowFlake.sequence = 0
	}

	snowFlake.lastTimestamp = timestamp

	return ((timestamp - epoch) << (timestampShift)) | (snowFlake.workerId << workerIdShift) | snowFlake.sequence
}

// 等待直到now>=snowFlake.lastTimestamp
func (snowFlake *SnowFlake) wait() int64 {
	timestamp := time.Now().UnixNano() / 1000000
	for ; timestamp < snowFlake.lastTimestamp; timestamp = time.Now().UnixNano() / 1000000 {
	}
	return timestamp
}

// 等待直到now>snowFlake.lastTimestamp
func (snowFlake *SnowFlake) waitNext() int64 {
	timestamp := time.Now().UnixNano() / 1000000
	for ; timestamp <= snowFlake.lastTimestamp; timestamp = time.Now().UnixNano() / 1000000 {
	}
	return timestamp
}
