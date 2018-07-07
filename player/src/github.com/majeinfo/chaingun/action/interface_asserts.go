package action

var _ Action = (*HttpAction)(nil)
var _ Action = (*SleepAction)(nil)
var _ Action = (*WSAction)(nil)
var _ Action = (*TcpAction)(nil)
var _ Action = (*UdpAction)(nil)
