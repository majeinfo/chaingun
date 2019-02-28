package action

var _ Action = (*HTTPAction)(nil)
var _ Action = (*SleepAction)(nil)
var _ Action = (*WSAction)(nil)
var _ Action = (*TCPAction)(nil)
var _ Action = (*UDPAction)(nil)
