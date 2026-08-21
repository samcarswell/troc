package test

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type E2eInstance struct {
	TrocExe   string
	NtfyToken string
	NtfyUri   string
}

var instance = E2eInstance{}

func GetInstance() E2eInstance {
	return instance
}

func ntfyInstance() (testcontainers.Container, error) {
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "binwiederhier/ntfy:v2.27",
			ExposedPorts: []string{"80/tcp"},
			Env: map[string]string{
				"NTFY_PASSWORD":                    "password",
				"NTFY_PASSWORD_HAS":                "$2a$13$QIItyPfXilSD5k2NHj58uOqM6vFYbiMztb5IwqKnFMISBZopIzfX.",
				"NTFY_VISITOR_REQUEST_LIMIT_BURST": "20000",
			},
			Cmd: []string{
				"serve",
			},
			WaitingFor: wait.ForListeningPort("80/tcp"),
		},
		Started:      true,
		ProviderType: 0,
		Logger:       nil,
		Reuse:        false,
	})
	if err != nil {
		return nil, err
	}
	err = container.Start(ctx)
	if err != nil {
		return nil, err
	}
	c, reader, err := container.Exec(ctx, []string{
		"/bin/sh",
		"-c",
		"/bin/mkdir /etc/ntfy && echo 'auth-file: \"/var/lib/ntfy/user.db\"' > /etc/ntfy/server.yml && echo 'auth-default-access: \"deny-all\"' >> /etc/ntfy/server.yml && /bin/mkdir /var/lib/ntfy && /bin/touch /var/lib/ntfy/user.db && ntfy user add --role=admin test",
	})
	if err != nil {
		return nil, err
	}
	buf := new(strings.Builder)
	_, err = io.Copy(buf, reader)
	if err != nil {
		return nil, err
	}
	if c != 0 {
		return nil, errors.New("exec returned non-zero exit code")
	}

	c, reader, err = container.Exec(ctx, []string{
		"ntfy",
		"token",
		"add",
		"test",
	})
	if err != nil {
		return nil, err
	}
	buf = new(strings.Builder)
	_, err = io.Copy(buf, reader)
	if err != nil {
		return nil, err
	}
	if c != 0 {
		return nil, errors.New("exec returned non-zero exit code")
	}
	instance.NtfyToken = strings.Split(buf.String(), " ")[1]

	instance.NtfyUri, err = container.PortEndpoint(ctx, "80", "http")
	if err != nil {
		return nil, err
	}

	return container, nil

}

func TestMain(m *testing.M) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	buildDir := path.Join(pwd, "..")
	dir, err := os.MkdirTemp(os.TempDir(), "build")
	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command("./build", dir)
	cmd.Dir = buildDir
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	instance.TrocExe = path.Join(dir, "troc")
	defer os.RemoveAll(dir)

	cont, err := ntfyInstance()
	if err != nil {
		panic(err)
	}

	exitVal := m.Run()
	cont.Terminate(context.Background())
	os.Exit(exitVal)
}
