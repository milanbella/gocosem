package gocosem

type AppConn struct {
	dconn             *DlmsConn
	applicationClient uint16
	logicalDevice     uint16
}
