// funcs
package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

//func BinarySearch(arr *[]int, leftIndex int, rightIndex int, findVal int) (int, int, bool) {
//	if leftIndex > rightIndex {
//		return 0, 0, false
//	}

//	//先找到中间下标
//	midddle := (leftIndex + rightIndex) / 2
//	if (*arr)[midddle] > findVal {
//		//说明要查找的数在左边  就应该向 leftIndex ---- (middle - 1)再次查找
//		rightIndex = midddle - 1
//		if rightIndex-leftIndex <= 1 {
//			return leftIndex, rightIndex, true
//		}
//		BinarySearch(arr, leftIndex, midddle-1, findVal)
//	} else if (*arr)[midddle] < findVal {
//		//如果 arr[middle] < findVal , 就应该向 middel+1---- rightIndex
//		leftIndex = midddle + 1
//		if rightIndex-leftIndex <= 1 {
//			return leftIndex, rightIndex, true
//		}
//		BinarySearch(arr, leftIndex, rightIndex, findVal)
//	} else {
//		return midddle, midddle, true
//	}

//}
//查表法计算热电偶测量的温度值
func CalcRdoWd(dy int) int {
	_, i, v, ok := BinarySearch(rdoVa, dy)
	switch {
	case ok:
		return i
	case i == 0, i == len(rdoVa)-1:
		return i
	default:
		aDy, bDy := dy-v, rdoVa[i+1]-v
		awd := float32(aDy) / float32(bDy)
		return i + int(awd)
	}
}

//二分法在dats数组中查找target值，返回值：查找次数,最接近目标值的数据下标,最接近的目标值,是否精确找到了目标值
func BinarySearch(dats []int, target int) (int, int, int, bool) {
	var l, m, r int = 0, 0, len(dats) - 1    //左、中、右下标
	var lv, mv, rv int = dats[0], 0, dats[r] //左、中、右数值
	var c int                                //下标差，差值
	//	fmt.Println("find.............", l, r, lv, rv, target)
	if target > rv {
		return 0, r, rv, false
	} else if target < lv {
		return 0, l, lv, false
	}
	for i := 0; ; i++ {
		c = r - l
		m = l + c/2
		mv, lv, rv = dats[m], dats[l], dats[r]
		//		fmt.Println(i, "try", l, m, r, "\t", lv, mv, target, rv)
		if c <= 1 {
			if c > 0 {
				//				fmt.Println(i, "find1", l, m, r, "\t", lv, mv, target, rv)
				return i, l, lv, false
			} else {
				//				fmt.Println(i, "find2", l, m, r, "\t", lv, mv, target, rv)
				return i, r, rv, false
			}
		}
		switch {
		case lv == target:
			//			fmt.Println(i, "find3", l, m, r, "\t", lv, mv, target, rv)
			return i, l, target, true
		case rv == target:
			//			fmt.Println(i, "find4", l, m, r, "\t", lv, mv, target, rv)
			return i, r, target, true
		case mv == target:
			//			fmt.Println(i, "find5", l, m, r, "\t", lv, mv, target, rv)
			return i, m, target, true
		case mv > target:
			r = m
		case mv < target:
			l = m
		}
	}
}

/**解析响应数组为数据项:	fjdz:1u,2u,4-,...
[数据代码:]<字节数>[类型代码]	=>	int|string00
类型代码,默认值u:小写字母:小端字节序;大写字母:大端字节序；
	'-'表示忽略此数据;i:有符号整形；u:无符号整形
**/
func DecodeDat(res []byte, datHeader string) map[string]interface{} {
	dat := make(map[string]interface{})
	var resAddr int //数据项地址
	ptn, _ := regexp.Compile("(?:([a-zA-Z0-9_]+):)?([1-9]+)([a-zA-Z-]?)")
	for i, hn := range strings.Split(datHeader, ",") {
		//		hn = strings.Trim(hn, ":")
		//数据项解码格式解析：数据项代码,数据项字节数,数据项类型
		dcd, dlen, dtype := fmt.Sprintf("d%02x", resAddr), 2, "u"
		if ps := ptn.FindStringSubmatch(hn); ps == nil {
			panic(fmt.Sprintf("第%d数据项格式错误:%s", i, hn))
		} else {
			if ps[1] != "" {
				dcd = ps[1]
			}
			dlen, _ = strconv.Atoi(ps[2])
			dtype = ps[3]
			//				dtype = []rune(ps[3])[0]
		}
		if dtype == "-" {
			resAddr += dlen
			continue
		}

		//数据解码
		switch dlen {
		case 1:
			if 0xfe == res[resAddr] {
				dat[dcd] = "--"
				break
			}
			//数据类型处理
			if strings.Index("iI", dtype) >= 0 {
				dat[dcd] = int(int8(res[resAddr]))
			} else { //if strings.Index("uU", dtype) >= 0
				dat[dcd] = int(uint8(res[resAddr]))
			}
			break

		case 2:
			lv, hv := res[resAddr], res[resAddr+1]
			if dtype >= "A" && dtype <= "Z" { //大端字节序
				lv, hv = hv, lv
			}
			if 0xfe == hv && 0xef == lv { //断线检测
				dat[dcd] = "--"
				break
			}
			//数据类型处理
			if strings.Index("iI", dtype) >= 0 {
				val := int16(uint16(hv)<<8 | uint16(lv))
				dat[dcd] = int(val)
			} else { //if strings.Index("uU", dtype) >= 0
				val := uint16(hv)<<8 | uint16(lv)
				dat[dcd] = int(val)
			}
		}

		log.Println(i, hn, dcd, dlen, dtype, resAddr, dat[dcd])

		resAddr += dlen
	}
	return dat
}

//握手： 地址  80  CRC低 CRC高
func Tx_CF0_req(addr byte) []byte {
	buf := []byte{addr, 0X80, 0, 0}
	crc16 := Crc16_rtu_A001(buf[:2])
	buf[2] = byte(crc16)
	buf[3] = byte(crc16 >> 8)
	return buf
}

//握手响应【地址|OK】检测，匹配符合true
func Tx_CF0_res_isMatch(addr byte, res []byte) bool {
	if res[0] == addr && res[1] == 'O' && res[2] == 'K' {
		return true
	}
	return false
}

//计算dat数组中offset开始的len字节的CRC16值，多项式 0xA001
func Crc16_rtu_A001(dat []byte) uint16 {
	var crc_reg, crc_gen uint32
	crc_reg, crc_gen = 0xFFFF, 0xA001

	for i, l := 0, len(dat); i < l; i++ {
		crc_reg = (uint32(dat[i]) & 0xff) ^ crc_reg
		for j := 8; j > 0; j-- {
			if crc_reg&0x01 == 1 {
				crc_reg >>= 1
				crc_reg ^= crc_gen
			} else {
				crc_reg >>= 1
			}
		}
	}
	return uint16(crc_reg)
}

//计算buf的crc16值并将值添加到末尾的2个字节:低位，高位
func SetCrc16(buf []byte) {
	bufLen := len(buf)
	src := buf[:bufLen-2]
	crc := Crc16_rtu_A001(src)
	buf[bufLen-2] = byte(crc)
	buf[bufLen-1] = byte(crc >> 8)
}

func ToUint16(high, low byte) uint16 {
	return uint16(high)<<8 | uint16(low)
}
