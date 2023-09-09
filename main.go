//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	usage = "Usage: mkmemkey [HKLM|HKCU|HKCR|HKU|HKCC]\\[KeyName]\nCreates a volatile Windows registry key (only exists in memory, doesn't persist a reboot)."

	_REG_OPTION_NON_VOLATILE = 0
	_REG_OPTION_VOLATILE     = 1
	_REG_OPENED_EXISTING_KEY = 2
	// ROOTKEY  [ HKLM | HKCU | HKCR | HKU | HKCC ]
	HKLM = registry.LOCAL_MACHINE
	HKCU = registry.CURRENT_USER
	HKU  = registry.USERS
	HKCC = registry.CURRENT_CONFIG
	HKCR = registry.CLASSES_ROOT
)

var (
	modadvapi32         = windows.NewLazySystemDLL("advapi32.dll")
	procRegCreateKeyExW = modadvapi32.NewProc("RegCreateKeyExW")
)

func regCreateKeyEx(key syscall.Handle, subkey *uint16, reserved uint32, class *uint16, options uint32, desired uint32, sa *syscall.SecurityAttributes, result *syscall.Handle, disposition *uint32) (regerrno error) {
	r0, _, _ := syscall.Syscall9(procRegCreateKeyExW.Addr(), 9, uintptr(key), uintptr(unsafe.Pointer(subkey)), uintptr(reserved), uintptr(unsafe.Pointer(class)), uintptr(options), uintptr(desired), uintptr(unsafe.Pointer(sa)), uintptr(unsafe.Pointer(result)), uintptr(unsafe.Pointer(disposition)))
	if r0 != 0 {
		regerrno = syscall.Errno(r0)
	}
	return
}

// CreateVolatileKey creates a volatile key named path under open key k.
// A volatile key exists only in memory and does not persist a reboot.
// CreateVolatileKey returns the new key and a boolean flag that reports
// whether the key already existed.
// The access parameter specifies the access rights for the key
// to be created.
func CreateVolatileKey(k registry.Key, path string, access uint32) (newk registry.Key, openedExisting bool, err error) {
	var h syscall.Handle
	var d uint32
	err = regCreateKeyEx(syscall.Handle(k), syscall.StringToUTF16Ptr(path),
		0, nil, _REG_OPTION_VOLATILE, access, nil, &h, &d)
	if err != nil {
		return 0, false, err
	}
	return registry.Key(h), d == _REG_OPENED_EXISTING_KEY, nil
}

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "ERROR - no RegKey given.")
		fmt.Fprintln(os.Stdout, usage)
		os.Exit(1)

	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(0)
	}

	path := os.Args[1]

	pathParts := strings.SplitN(path, `\`, 2)

	rootKey, err := func(rootKeyCandidate string) (registry.Key, error) {
		switch pathParts[0] {
		case "HKLM":
			return HKLM, nil
		case "HKEY_LOCAL_MACHINE":
			return HKLM, nil
		case "HKCU":
			return HKCU, nil
		case "HKEY_CURRENT_USER":
			return HKCU, nil
		case "HKCR":
			return HKCR, nil
		case "HKEY_CLASSES_ROOT":
			return HKCR, nil
		case "HKU":
			return HKU, nil
		case "HKEY_USERS":
			return HKU, nil
		case "HKCC":
			return HKCC, nil
		case "HKEY_CURRENT_CONFIG":
			return HKCC, nil
		default:
			return *new(registry.Key), errors.New("Invalid root key. Must be one of: [ HKLM | HKCU | HKCR | HKU | HKCC ].")

		}

	}(pathParts[0])

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, existing, err := CreateVolatileKey(rootKey, pathParts[1], registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if existing {
		fmt.Fprintln(os.Stderr, "WARNING - registry key already existed, no change made")
		os.Exit(2)
	}
	fmt.Fprintln(os.Stdout, "Success.")
	os.Exit(0)

}
