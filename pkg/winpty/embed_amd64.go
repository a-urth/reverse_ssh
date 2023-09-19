//go:build windows
// +build windows

package winpty

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func writeBinaries() error {

	vsn := windows.RtlGetVersion()

	/*
		https://msdn.microsoft.com/en-us/library/ms724832(VS.85).aspx
		Windows 10					10.0*
		Windows Server 2016			10.0*
		Windows 8.1					6.3*
		Windows Server 2012 R2		6.3*
		Windows 8					6.2
		Windows Server 2012			6.2
		Windows 7					6.1
		Windows Server 2008 R2		6.1
		Windows Server 2008			6.0
		Windows Vista				6.0
		Windows Server 2003 R2		5.2
		Windows Server 2003			5.2
		Windows XP 64-Bit Edition	5.2
		Windows XP					5.1
		Windows 2000				5.0
	*/

	dllType := "regular"
	if vsn.MajorVersion == 5 {
		dllType = "xp"
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Println("unable to get cache directory for writing winpty pe's writing may fail if directory is ro")
	}

	if err == nil {
		winptyDllName = cacheDir + "\\temp\\" + winptyDllName
		winptyAgentName = cacheDir + "\\temp\\" + winptyAgentName
	}

	if _, err := os.Stat(winptyDllName); os.IsNotExist(err) {
		dll, err := Asset(path.Join("embed", "x64", dllType, "winpty.dll"))
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(winptyDllName, dll, 0700)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(winptyAgentName); os.IsNotExist(err) {
		dll, err := Asset(path.Join("embed", "x64", dllType, "winpty-agent.exe"))
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(winptyAgentName, dll, 0700)
		if err != nil {
			return err
		}
	}

	return nil
}

func createAgentCfg(flags uint32) (uintptr, error) {
	var errorPtr uintptr

	if winpty_error_free == nil {
		return uintptr(0), errors.New("winpty was not initalised")
	}

	err := winpty_error_free.Find() // check if dll available
	if err != nil {
		return uintptr(0), err
	}

	defer winpty_error_free.Call(errorPtr)

	agentCfg, _, _ := winpty_config_new.Call(uintptr(flags), uintptr(unsafe.Pointer(errorPtr)))
	if agentCfg == uintptr(0) {
		return 0, fmt.Errorf("Unable to create agent config, %s", GetErrorMessage(errorPtr))
	}

	return agentCfg, nil
}

func createSpawnCfg(flags uint32, appname, cmdline, cwd string, env []string) (uintptr, error) {
	var errorPtr uintptr
	defer winpty_error_free.Call(errorPtr)

	cmdLineStr, err := syscall.UTF16PtrFromString(cmdline)
	if err != nil {
		return 0, fmt.Errorf("Failed to convert cmd to pointer.")
	}

	appNameStr, err := syscall.UTF16PtrFromString(appname)
	if err != nil {
		return 0, fmt.Errorf("Failed to convert app name to pointer.")
	}

	cwdStr, err := syscall.UTF16PtrFromString(cwd)
	if err != nil {
		return 0, fmt.Errorf("Failed to convert working directory to pointer.")
	}

	envStr, err := UTF16PtrFromStringArray(env)

	if err != nil {
		return 0, fmt.Errorf("Failed to convert cmd to pointer.")
	}

	spawnCfg, _, _ := winpty_spawn_config_new.Call(
		uintptr(flags),
		uintptr(unsafe.Pointer(appNameStr)),
		uintptr(unsafe.Pointer(cmdLineStr)),
		uintptr(unsafe.Pointer(cwdStr)),
		uintptr(unsafe.Pointer(envStr)),
		uintptr(unsafe.Pointer(errorPtr)),
	)

	if spawnCfg == uintptr(0) {
		return 0, fmt.Errorf("Unable to create spawn config, %s", GetErrorMessage(errorPtr))
	}

	return spawnCfg, nil
}
