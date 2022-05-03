package domain

type Connection interface {
	Open()
	Close()
	IsOpen() bool
}

type ConnectionsStorage interface {
	GetConnection(ipAddress int32) Connection
	OnNewRemoteConnection(remotePeer int32, conn Connection)
	Run()
	Shutdown()
}
