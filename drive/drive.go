package drive

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"google.golang.org/api/drive/v3"
)

type stringSyncMap struct {
	sm sync.Map
}

func (f *stringSyncMap) Lookup(k string) string {
	v, ok := f.sm.Load(k)
	if !ok {
		return ""
	}

	t, ok := v.(string)
	if !ok {
		fmt.Printf("Lookup %s failed: type is not string\n", k)
		return ""
	}

	return t
}

func (f *stringSyncMap) Store(k, id string) {
	f.sm.Store(k, id)
}

type FilePicker struct {
	drive         *drive.Service
	cacheIDByPath stringSyncMap
}

func NewFilePicker(client *http.Client) (*FilePicker, error) {
	srv, err := drive.New(client)
	if err != nil {
		return nil, err
	}
	return &FilePicker{
		drive:         srv,
		cacheIDByPath: stringSyncMap{},
	}, nil
}

// Lookup Get fileID from path
func (p *FilePicker) Lookup(path string) (string, error) {
	currentID := "root"
	names := strings.Split(path, "/")[1:]
	for i, name := range names {
		currentPath := "/" + strings.Join(names[:i+1], "/")
		cachedID := p.cacheIDByPath.Lookup(currentPath)
		if cachedID != "" {
			currentID = cachedID
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
			p.cacheIDByPath.Store(currentPath, currentID)
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
