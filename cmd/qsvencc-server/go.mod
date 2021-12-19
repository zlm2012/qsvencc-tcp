module qsvencc-server

go 1.17

replace local.packages/qsvencc-tcp => ../..

require (
	github.com/google/uuid v1.3.0
	local.packages/qsvencc-tcp v0.0.0-00010101000000-000000000000
)
