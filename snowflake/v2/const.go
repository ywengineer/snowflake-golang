package v2

import "sync"

/**
 * 起始的时间戳
 */
const START_STAMP uint64 = 1566554916249

/**
 * 每一部分占用的位数
 */
const SEQUENCE_BIT uint64 = 12   //序列号占用的位数
const MACHINE_BIT uint64 = 5     //机器标识占用的位数
const DATA_CENTER_BIT uint64 = 5 //数据中心占用的位数
/**
 * 每一部分的最大值
 */
const MAX_DATA_CENTER_NUM uint64 = ^(-1 << DATA_CENTER_BIT)
const MAX_MACHINE_NUM uint64 = ^(-1 << MACHINE_BIT)
const MAX_SEQUENCE uint64 = ^(-1 << SEQUENCE_BIT)

/**
 * 每一部分向左的位移
 */
const MACHINE_LEFT = SEQUENCE_BIT
const DATA_CENTER_LEFT = SEQUENCE_BIT + MACHINE_BIT
const TIMESTAMP_LEFT = DATA_CENTER_LEFT + DATA_CENTER_BIT

var cmu = sync.Mutex{}
