// Code generated by 'go generate'; DO NOT EDIT.

package winwfp

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var _ unsafe.Pointer

// Do the interface allocations only once for common
// Errno values.
const (
	errnoERROR_IO_PENDING = 997
)

var (
	errERROR_IO_PENDING error = syscall.Errno(errnoERROR_IO_PENDING)
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	// TODO: add more here, after collecting data on the common
	// error values see on Windows. (perhaps when running
	// all.bat?)
	return e
}

var (
	modfwpuclnt = windows.NewLazySystemDLL("fwpuclnt.dll")
	modKernel32 = windows.NewLazySystemDLL("Kernel32.dll")

	procConvertSidToStringSidW    = modfwpuclnt.NewProc("ConvertSidToStringSidW")
	procAllocateAndInitializeSid  = modfwpuclnt.NewProc("AllocateAndInitializeSid")
	procFreeSid                   = modfwpuclnt.NewProc("FreeSid")
	procFwpmEngineOpen0           = modfwpuclnt.NewProc("FwpmEngineOpen0")
	procFwpmEngineClose0          = modfwpuclnt.NewProc("FwpmEngineClose0")
	procFwpmSubLayerAdd0          = modfwpuclnt.NewProc("FwpmSubLayerAdd0")
	procGetModuleFileNameW        = modKernel32.NewProc("GetModuleFileNameW")
	procFwpmGetAppIdFromFileName0 = modfwpuclnt.NewProc("FwpmGetAppIdFromFileName0")
	procFwpmFilterAdd0            = modfwpuclnt.NewProc("FwpmFilterAdd0")
)

func convertSidToStringSidW(Sid *wtSid, StringSid *uint16) (result uint8) {
	r0, _, _ := syscall.Syscall(procConvertSidToStringSidW.Addr(), 2, uintptr(unsafe.Pointer(Sid)), uintptr(unsafe.Pointer(StringSid)), 0)
	result = uint8(r0)
	return
}

func allocateAndInitializeSid(pIdentifierAuthority *SidIdentifierAuthority, nSubAuthorityCount uint8, nSubAuthority0 uint32, nSubAuthority1 uint32, nSubAuthority2 uint32, nSubAuthority3 uint32, nSubAuthority4 uint32, nSubAuthority5 uint32, nSubAuthority6 uint32, nSubAuthority7 uint32, pSid unsafe.Pointer) (result uint8) {
	r0, _, _ := syscall.Syscall12(procAllocateAndInitializeSid.Addr(), 11, uintptr(unsafe.Pointer(pIdentifierAuthority)), uintptr(nSubAuthorityCount), uintptr(nSubAuthority0), uintptr(nSubAuthority1), uintptr(nSubAuthority2), uintptr(nSubAuthority3), uintptr(nSubAuthority4), uintptr(nSubAuthority5), uintptr(nSubAuthority6), uintptr(nSubAuthority7), uintptr(pSid), 0)
	result = uint8(r0)
	return
}

func freeSid(Sid *wtSid) (result uint8) {
	r0, _, _ := syscall.Syscall(procFreeSid.Addr(), 1, uintptr(unsafe.Pointer(Sid)), 0, 0)
	result = uint8(r0)
	return
}

func fwpmEngineOpen0(serverName *uint16, authnService wtRpcCAuthN, authIdentity *wtSecWinntAuthIdentityW, session *wtFwpmSession0, engineHandle unsafe.Pointer) (result uint) {
	r0, _, _ := syscall.Syscall6(procFwpmEngineOpen0.Addr(), 5, uintptr(unsafe.Pointer(serverName)), uintptr(authnService), uintptr(unsafe.Pointer(authIdentity)), uintptr(unsafe.Pointer(session)), uintptr(engineHandle), 0)
	result = uint(r0)
	return
}

func fwpmEngineClose0(engineHandle uintptr) (result uint32) {
	r0, _, _ := syscall.Syscall(procFwpmEngineClose0.Addr(), 1, uintptr(engineHandle), 0, 0)
	result = uint32(r0)
	return
}

func fwpmSubLayerAdd0(engineHandle uintptr, subLayer *wtFwpmSublayer0, sd uintptr) (result uint32) {
	r0, _, _ := syscall.Syscall(procFwpmSubLayerAdd0.Addr(), 3, uintptr(engineHandle), uintptr(unsafe.Pointer(subLayer)), uintptr(sd))
	result = uint32(r0)
	return
}

func getModuleFileNameW(hModule uintptr, lpFilename *uint16, nSize uint32) (result uint32) {
	r0, _, _ := syscall.Syscall(procGetModuleFileNameW.Addr(), 3, uintptr(hModule), uintptr(unsafe.Pointer(lpFilename)), uintptr(nSize))
	result = uint32(r0)
	return
}

func fwpmGetAppIdFromFileName0(fileName *uint16, appId unsafe.Pointer) (result uint32) {
	r0, _, _ := syscall.Syscall(procFwpmGetAppIdFromFileName0.Addr(), 2, uintptr(unsafe.Pointer(fileName)), uintptr(appId), 0)
	result = uint32(r0)
	return
}

func fwpmFilterAdd0(engineHandle uintptr, filter *wtFwpmFilter0, sd uintptr, id *uint64) (result uint32) {
	r0, _, _ := syscall.Syscall6(procFwpmFilterAdd0.Addr(), 4, uintptr(engineHandle), uintptr(unsafe.Pointer(filter)), uintptr(sd), uintptr(unsafe.Pointer(id)), 0, 0)
	result = uint32(r0)
	return
}
