package common

const (
	BEGIN_SEND_BETS       = "BEGIN"
	END_SEND_BETS         = "END"
	BATCH_SENT_OK         = "OK"
	BATCH_SENT_FAIL       = "FAIL"
	BATCH_END_CHAR   byte = '|'
	MAX_BATCH_SIZE        = 8192
	READ_BUFFER_SIZE      = 1024
)
