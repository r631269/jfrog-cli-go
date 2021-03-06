package docker

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"io"
	"os/exec"
	"path"
	"strings"
)

func New(imageTag string) Image {
	return &image{tag: imageTag}
}

// Docker image
type Image interface {
	Push() error
	Id() (string, error)
	ParentId() (string, error)
	Tag() string
	Path() string
	Name() string
}

// Internal implementation of docker image
type image struct {
	tag string
}

// Push docker image
func (image *image) Push() error {
	cmd := &pushCmd{image: image}
	return utils.RunCmd(cmd)
}

// Get docker image tag
func (image *image) Tag() string {
	return image.tag
}

// Get docker image ID
func (image *image) Id() (string, error) {
	cmd := &getImageIdCmd{image: image}
	content, err := utils.RunCmdOutput(cmd)
	return strings.Trim(string(content), "\n"), err
}

// Get docker parent image ID
func (image *image) ParentId() (string, error) {
	cmd := &getParentId{image: image}
	content, err := utils.RunCmdOutput(cmd)
	return strings.Trim(string(content), "\n"), err
}

// Get docker image relative path in Artifactory
func (image *image) Path() string {
	indexOfFirstSlash := strings.Index(image.tag, "/")
	indexOfLastColon := strings.LastIndex(image.tag, ":")

	if indexOfLastColon < 0 || indexOfLastColon < indexOfFirstSlash {
		return path.Join(image.tag[indexOfFirstSlash:], "latest")
	}
	return path.Join(image.tag[indexOfFirstSlash:indexOfLastColon], image.tag[indexOfLastColon+1:])
}

// Get docker image name
func (image *image) Name() string {
	indexOfLastSlash := strings.LastIndex(image.tag, "/")
	indexOfLastColon := strings.LastIndex(image.tag, ":")

	if indexOfLastColon < 0 || indexOfLastColon < indexOfLastSlash {
		return image.tag[indexOfLastSlash+1:] + ":latest"
	}
	return image.tag[indexOfLastSlash+1:]
}

// Image push command
type pushCmd struct {
	image *image
}

func (pushCmd *pushCmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "docker")
	cmd = append(cmd, "push")
	cmd = append(cmd, pushCmd.image.tag)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (pushCmd *pushCmd) GetEnv() map[string]string {
	return map[string]string{}
}

func (pushCmd *pushCmd) GetStdWriter() io.WriteCloser {
	return nil
}
func (pushCmd *pushCmd) GetErrWriter() io.WriteCloser {
	return nil
}

// Image get image id command
type getImageIdCmd struct {
	image *image
}

func (getImageId *getImageIdCmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "docker")
	cmd = append(cmd, "images")
	cmd = append(cmd, "--format", "{{.ID}}")
	cmd = append(cmd, "--no-trunc")
	cmd = append(cmd, getImageId.image.tag)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (getImageId *getImageIdCmd) GetEnv() map[string]string {
	return map[string]string{}
}

func (getImageId *getImageIdCmd) GetStdWriter() io.WriteCloser {
	return nil
}

func (getImageId *getImageIdCmd) GetErrWriter() io.WriteCloser {
	return nil
}

// Image get parent image id command
type getParentId struct {
	image *image
}

func (getImageId *getParentId) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "docker")
	cmd = append(cmd, "inspect")
	cmd = append(cmd, "--format", "{{.Parent}}")
	cmd = append(cmd, getImageId.image.tag)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (getImageId *getParentId) GetEnv() map[string]string {
	return map[string]string{}
}

func (getImageId *getParentId) GetStdWriter() io.WriteCloser {
	return nil
}

func (getImageId *getParentId) GetErrWriter() io.WriteCloser {
	return nil
}

// Get docker registry from tag
func ResolveRegistryFromTag(imageTag string) (string, error) {
	indexOfFirstSlash := strings.Index(imageTag, "/")
	if indexOfFirstSlash < 0 {
		err := errorutils.CheckError(errors.New("Invalid image tag received for pushing to Artifactory - tag does not include a slash."))
		return "", err
	}

	indexOfSecondSlash := strings.Index(imageTag[indexOfFirstSlash+1:], "/")
	// Reverse proxy Artifactory
	if indexOfSecondSlash < 0 {
		return imageTag[:indexOfFirstSlash], nil
	}
	// Can be reverse proxy or proxy-less Artifactory
	indexOfSecondSlash += indexOfFirstSlash + 1
	return imageTag[:indexOfSecondSlash], nil
}

// Login command
type LoginCmd struct {
	DockerRegistry string
	Username       string
	Password       string
}

func (loginCmd *LoginCmd) GetCmd() *exec.Cmd {
	cmdLogin := "| docker login " + loginCmd.DockerRegistry + " --username=" + loginCmd.Username + " --password-stdin"

	if cliutils.IsWindows() {
		cmd := "echo %DOCKER_PASS%" + cmdLogin
		return exec.Command("cmd", "/C", cmd)
	}

	cmd := "echo $DOCKER_PASS " + cmdLogin
	return exec.Command("bash", "-c", cmd)
}

func (loginCmd *LoginCmd) GetEnv() map[string]string {
	return map[string]string{"DOCKER_PASS": loginCmd.Password}
}

func (loginCmd *LoginCmd) GetStdWriter() io.WriteCloser {
	return nil
}

func (loginCmd *LoginCmd) GetErrWriter() io.WriteCloser {
	return nil
}
