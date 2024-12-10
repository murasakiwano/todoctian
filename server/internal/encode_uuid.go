package internal

import "encoding/hex"

// This function takes a bytes array and encodes them as a UUID string.
// Since the String() method has not been released in a new version of pgtype
// as of 2024-12-10, I wrote it here. See https://github.com/jackc/pgx/blob/master/pgtype/uuid.go
// for more info.
func EncodeUUID(src [16]byte) string {
	var buf [36]byte

	hex.Encode(buf[0:8], src[:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], src[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], src[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], src[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], src[10:])

	return string(buf[:])
}
