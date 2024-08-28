package ssh_utils

import (
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHParams struct {
	User    string
	Key     []byte
	Hosts   []string
	Port    int
	Command string
}

// Connect to host and return client, session and error if any.
func ConnectToHost(user, host string, key []byte) (*ssh.Client, *ssh.Session, error) {
	sshConfig := &ssh.ClientConfig{
		User:    user,
		Auth:    []ssh.AuthMethod{PublicKeyFile(key)},
		Timeout: time.Second * 3,
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		_ = client.Close()
		return nil, nil, err
	}

	return client, session, nil
}

// Parse a private key file and return ssh auth method.
func PublicKeyFile(buffer []byte) ssh.AuthMethod {
	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

// Connect to host, run command and get output.
func GetOutput(user, host, command string, key []byte) (string, error) {
	client, session, err := ConnectToHost(user, host, key)
	if err != nil {
		return "", err
	}
	out, err := session.CombinedOutput(command)
	if err != nil {
		return "", err
	}
	_ = client.Close()
	return string(out), nil
}
