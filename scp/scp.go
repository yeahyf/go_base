///scp接口封装
package scp

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"time"

	"github.com/yeahyf/go_base/log"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

//不使用密码，使用密钥
func ScpFileForKey(host, localFilePath, remoteDir, user, key string, port int) error {
	return scpFile(host, localFilePath, remoteDir, user, "", key, port)
}

//host 主机ip地址 localFilePath 本地文件路径 remoteDir远程路径,兼容老的方法
func ScpFile(host, localFilePath, remoteDir, user, passwd string, port int) error {
	return scpFile(host, localFilePath, remoteDir, user, passwd, "", port)
}

func scpFile(host, localFilePath, remoteDir, user, passwd, key string, port int) error {
	// 这里换成实际的 SSH 连接的 用户名，密码，主机名或IP，SSH端口
	sftpClient, err := sftpConnect(user, passwd, host, key, port)
	if err != nil {
		log.Error(err)
		return err
	}
	defer sftpClient.Close()

	// 用来测试的本地文件路径 和 远程机器上的文件夹
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		log.Error(err)
		return err
	}
	defer srcFile.Close()

	var remoteFileName = path.Base(localFilePath)
	dstFile, err := sftpClient.Create(path.Join(remoteDir, remoteFileName))
	if err != nil {
		log.Error(err)
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		log.Error(err)
		log.Errorf("Copy File to %s path: %s  Failed!", host, remoteDir)
		return err
	}
	if log.IsDebug() {
		log.Debugf("Copy File to %s path: %s  Success!", host, remoteDir)
	}
	return nil
}

//构建一个SSH客户端
func sshClient(user, password, host, key string, port int, cipherList []string) (*ssh.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		config       ssh.Config
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0, 1)
	//是否存在密钥
	if key == "" {
		auth = append(auth, ssh.Password(password))
	} else {
		pemBytes, err := ioutil.ReadFile(key)
		if err != nil {
			return nil, err
		}

		var signer ssh.Signer
		if password == "" {
			signer, err = ssh.ParsePrivateKey(pemBytes)
		} else {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(password))
		}
		if err != nil {
			return nil, err
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	if len(cipherList) == 0 {
		config = ssh.Config{
			Ciphers: []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"},
		}
	} else {
		config = ssh.Config{
			Ciphers: cipherList,
		}
	}

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		Config:  config,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	return client, nil
}

//构建一个SSH的终端会话
func sshSession(user, password, host, key string, port int) (*ssh.Session, error) {
	client, err := sshClient(user, password, host, key, port, nil)
	if err != nil {
		return nil, err
	}

	// create session
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}

	//终端模式
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return nil, err
	}

	return session, nil
}

//构建一个Sftp的客户端
func sftpConnect(user, password, host, key string, port int) (*sftp.Client, error) {
	sess, err := sshClient(user, password, host, key, port, nil)
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(sess)
	if err != nil {
		return nil, err
	}
	return sftpClient, nil
}
