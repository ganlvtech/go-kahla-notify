package toast

import (
	"os/exec"
)

type SnoreToast struct {
	Path string
}

func New(path string) *SnoreToast {
	if path == "" {
		path = "SnoreToast.exe"
	}
	return &SnoreToast{Path: path}
}

func (s *SnoreToast) Toast(title string, message string) error {
	err := exec.Command(s.Path, "-t", title, "-m", message, "-s", "Notification.IM").Run()
	if err != nil {
		return err
	}
	return nil
}

func (s *SnoreToast) ToastWithImage(title string, message string, imagePath string) error {
	err := exec.Command(s.Path, "-t", title, "-m", message, "-p", imagePath, "-s", "Notification.IM").Run()
	if err != nil {
		return err
	}
	return nil
}
