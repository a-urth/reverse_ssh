//go:build windows
// +build windows

package winpty

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"unsafe"
)

func writeBinaries() error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Println("unable to get cache directory for writing winpty pe's writing may fail if directory is ro")
	}

	if err == nil {
		winptyDllName = cacheDir + "\\temp\\" + winptyDllName
		winptyAgentName = cacheDir + "\\temp\\" + winptyAgentName
	}

	if _, err := os.Stat(winptyDllName); os.IsNotExist(err) {
		dll, err := Asset("embed/winpty.dll")
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(winptyDllName, dll, 0700)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(winptyAgentName); os.IsNotExist(err) {
		dll, err := Asset("embed/winpty-agent.exe")
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
