package action

var _ Action = (*HTTPAction)(nil)
var _ Action = (*SleepAction)(nil)
var _ Action = (*WSAction)(nil)
var _ Action = (*TCPAction)(nil)
var _ Action = (*UDPAction)(nil)
var _ Action = (*MQTTAction)(nil)
var _ Action = (*MongoDBAction)(nil)
var _ Action = (*SQLAction)(nil)
