package halloween

type HalloweenServer struct {
	Join      chan JoinMessage
	Leave     chan LeaveMessage
	Broadcast chan BroadcastMassage
}
