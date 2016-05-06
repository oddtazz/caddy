package caddy

import (
	"log"
	"os"
)

// File descriptors for in-process restart
var restartFds map[string]*os.File = make(map[string]*os.File)

// restartInProc restarts Caddy forcefully in process using newCaddyfile.
func restartInProc(newCaddyfile Input) error {
	wg.Add(1) // Barrier so Wait() doesn't unblock
	defer wg.Done()

	// Add file descriptors of all the sockets for new instance
	serversMu.Lock()
	for _, s := range servers {
		restartFds[s.Addr] = s.ListenerFd()
	}
	serversMu.Unlock()

	caddyfileMu.Lock()
	oldCaddyfile := caddyfile
	caddyfileMu.Unlock()

	if stopErr := Stop(); stopErr != nil {
		return stopErr
	}

	err := Start(newCaddyfile)
	if err != nil {
		// Revert to old Caddyfile
		if oldErr := Start(oldCaddyfile); oldErr != nil {
			log.Printf("[ERROR] Restart: in-process restart failed and cannot revert to old Caddyfile: %v", oldErr)
		}
	}

	return err
}
