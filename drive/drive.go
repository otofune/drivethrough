package drive

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"google.golang.org/api/drive/v3"
)

type FilePicker struct {
	drive         *drive.Service
	cacheIDByPath map[string]string
}

func NewFilePicker(client *http.Client) (*FilePicker, error) {
	srv, err := drive.New(client)
	if err != nil {
		return nil, err
	}
	return &FilePicker{
		drive:         srv,
		cacheIDByPath: map[string]string{},
	}, nil
}

// Lookup Get fileID from path
func (p *FilePicker) Lookup(path string) (string, error) {
	currentID := "root"
	names := strings.Split(path, "/")[1:]
	for i, name := range names {
		currentPath := "/" + strings.Join(names[:i+1], "/")
		if p.cacheIDByPath[currentPath] != "" {
			currentID = p.cacheIDByPath[currentPath]
			continue
		}

		if name == "" {
			return "", fmt.Errorf("fuck empty string")
		}
		isDirectory := i != len(names)-1

		query := fmt.Sprintf("name = '%s' and '%s' in parents", strings.ReplaceAll(name, ",", "\\,"), currentID)
		if isDirectory {
			query += " and mimeType = 'application/vnd.google-apps.folder'"
		}

		list, err := p.drive.Files.List().Fields("files(id, name, mimeType, shortcutDetails)").Q(query).Do()
		if err != nil {
			return "", err
		}

		if len(list.Files) == 0 {
			return "", fmt.Errorf("ENOENT %s (parent:%s)", currentPath, currentID)
		}
		if len(list.Files) == 1 {
			file := list.Files[0]
			currentID = file.Id
			if file.MimeType == "application/vnd.google-apps.shortcut" && file.ShortcutDetails != nil {
				// FIXME: map で持つ構造がショートカットを前提にしていないので、そこが間違っている気がする
				currentID = file.ShortcutDetails.TargetId
			}
			p.cacheIDByPath[currentPath] = currentID
			continue
		}

		// debug
		for _, i := range list.Files {
			fmt.Printf("Candidate(%s): %s (%s)\n", currentPath, i.Name, i.Id)
		}
		return "", fmt.Errorf("filename is duplicated")
	}
	return currentID, nil
}

func (p *FilePicker) Read(fileID string) (io.ReadCloser, error) {
	resp, err := p.drive.Files.Get(fileID).Download()
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
