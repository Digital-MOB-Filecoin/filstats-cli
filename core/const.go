package core

import "time"

// name of the file to store the token
const TokenFile = "token.dat"

// Intervals at which to poll and send various telemetry info
const (
	HeartbeatInterval           = 15 * time.Second
	PeersInterval               = 15 * time.Second
	MpoolSizeInterval           = 15 * time.Second
	SyncingInterval             = 15 * time.Second
	NetworkStoragePowerInterval = 10 * time.Minute
)
