package sync

import (
	"bookbox-backend/pkg/logger"
	"time"

	"github.com/pkg/sftp"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

func ConnectToFtp(ftpUrl string) (sftpClient *sftp.Client, err error) {
	// SSH client config
	config := &ssh.ClientConfig{
		User: "bztr7423",
		Auth: []ssh.AuthMethod{
			ssh.Password("ZZbwMcqg0h4rNcpeQGCI"),
		},
		// This is to accept any host key.
		// Consider using ssh.FixedHostKey() for higher security
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	// Connect to the SSH server
	conn, err := ssh.Dial("tcp", ftpUrl, config)
	if err != nil {
		logger.Log.Error("failed to dial", zap.Error(err))
		return
	}

	// Create a new SFTP client
	sftpClient, err = sftp.NewClient(conn)
	if err != nil {
		logger.Log.Error("failed to create sftp client", zap.Error(err))
		conn.Close()
		return
	}

	return
}
