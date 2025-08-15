package internal

import "encoding/hex"

/*
*

协议

字段	    大小	    描述
Header	1 字节	固定值 0xA8
Type	1 字节	操作类型 (0x01-0x04)
Length	1 字节	内容长度 (0-255)
Content	0-255字节	UTF-8 编码的内容数据
CRC8	1 字节	CRC8-CCITT 校验和

示例: A8020033
*/

const (
	HeaderByte       byte = 0xA8
	TypeByteReboot   byte = 0x01
	TypeByteClear    byte = 0x02
	TypeByteShutdown byte = 0x03
	minFrameLen           = 4 // Header + Type + Length + CRC (Length 可为 0)
)

var validTypes = map[byte]struct{}{
	TypeByteReboot:   {},
	TypeByteClear:    {},
	TypeByteShutdown: {},
}

// parseDeviceFrame 解析设备指令帧，返回 (Type, Content, ok)
func parseDeviceFrame(frame []byte) (typ byte, content []byte, ok bool) {
	if len(frame) < minFrameLen {
		return
	}
	if frame[0] != HeaderByte {
		return
	}
	typ = frame[1]
	if _, okType := validTypes[typ]; !okType {
		return
	}
	contentLen := int(frame[2])
	if contentLen != len(frame)-minFrameLen {
		return
	}
	if crc8(frame[:len(frame)-1]) != frame[len(frame)-1] {
		return
	}
	if contentLen > 0 {
		content = frame[3 : 3+contentLen]
	}
	ok = true
	return
}

// IsDeviceQR 是否为任一合法设备指令
func IsDeviceQR(frame []byte) bool {
	_, _, ok := parseDeviceFrame(frame)
	return ok
}

// ---- 按类型判断 ----

func IsRebootQR(frame []byte) bool   { return isTypeFrame(frame, TypeByteReboot) }
func IsClearQR(frame []byte) bool    { return isTypeFrame(frame, TypeByteClear) }
func IsShutdownQR(frame []byte) bool { return isTypeFrame(frame, TypeByteShutdown) }

func IsDeviceQRString(s string) bool {
	b, err := hex.DecodeString(s)
	if err != nil {
		return false
	}
	return IsDeviceQR(b)
}
func IsRebootQRString(s string) bool {
	b, err := hex.DecodeString(s)
	if err != nil {
		return false
	}
	return IsRebootQR(b)
}
func IsClearQRString(s string) bool {
	b, err := hex.DecodeString(s)
	if err != nil {
		return false
	}
	return IsClearQR(b)
}
func IsShutdownQRString(s string) bool {
	b, err := hex.DecodeString(s)
	if err != nil {
		return false
	}
	return IsShutdownQR(b)
}

// 内部：仅在类型匹配时再做长度与 CRC 校验
func isTypeFrame(frame []byte, wantType byte) bool {
	if len(frame) < minFrameLen {
		return false
	}
	if frame[0] != HeaderByte || frame[1] != wantType {
		return false
	}
	contentLen := int(frame[2])
	if contentLen != len(frame)-minFrameLen {
		return false
	}
	return crc8(frame[:len(frame)-1]) == frame[len(frame)-1]
}

// CRC8 (poly 0x07, init 0x00) 与JavaScript版本对齐
func crc8(data []byte) byte {
	var crc byte = 0x00
	polynomial := byte(0x07)

	for _, b := range data {
		crc ^= b
		for bit := 0; bit < 8; bit++ {
			if crc&0x80 != 0 {
				crc = ((crc << 1) ^ polynomial) & 0xFF
			} else {
				crc = (crc << 1) & 0xFF
			}
		}
	}
	return crc
}
