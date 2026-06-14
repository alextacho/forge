package tree

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
)

type Kind string

const (
	KindAbsent    Kind = "absent"
	KindFile      Kind = "file"
	KindDirectory Kind = "directory"
	KindSymlink   Kind = "symlink"
)

type Inspection struct {
	Path  string
	Kind  Kind
	Mode  fs.FileMode
	Nodes []Node
}

type Node struct {
	Path string      `json:"path"`
	Kind Kind        `json:"kind"`
	Mode fs.FileMode `json:"mode"`
}

func Inspect(root *os.Root, name string) (Inspection, error) {
	if err := validateParentComponents(root, name); err != nil {
		return Inspection{}, err
	}
	info, err := root.Lstat(name)
	if errors.Is(err, os.ErrNotExist) {
		return Inspection{Path: name, Kind: KindAbsent}, nil
	}
	if err != nil {
		return Inspection{}, fmt.Errorf("inspect %q: %w", name, err)
	}
	kind, err := kindOf(info.Mode())
	if err != nil {
		return Inspection{}, fmt.Errorf("inspect %q: %w", name, err)
	}
	result := Inspection{
		Path: name,
		Kind: kind,
		Mode: info.Mode().Perm(),
		Nodes: []Node{{
			Path: ".",
			Kind: kind,
			Mode: info.Mode().Perm(),
		}},
	}
	if kind != KindDirectory {
		return result, nil
	}

	err = fs.WalkDir(root.FS(), name, func(entryPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entryPath == name {
			return nil
		}
		info, err := root.Lstat(entryPath)
		if err != nil {
			return err
		}
		if _, err := kindOf(info.Mode()); err != nil {
			return fmt.Errorf("%q: %w", entryPath, err)
		}
		result.Nodes = append(result.Nodes, Node{
			Path: strings.TrimPrefix(entryPath, name+"/"),
			Kind: mustKind(info.Mode()),
			Mode: info.Mode().Perm(),
		})
		return nil
	})
	if err != nil {
		return Inspection{}, fmt.Errorf("inspect tree %q: %w", name, err)
	}
	return result, nil
}

func Copy(source *os.Root, sourceName string, destination *os.Root, destinationName string) error {
	return copyTree(source, sourceName, destination, destinationName, false)
}

func CopyForStorage(source *os.Root, sourceName string, destination *os.Root, destinationName string) error {
	return copyTree(source, sourceName, destination, destinationName, true)
}

func ValidateCopySource(root *os.Root, name string) (Inspection, error) {
	inspection, err := Inspect(root, name)
	if err != nil || inspection.Kind == KindAbsent {
		return inspection, err
	}
	for _, node := range inspection.Nodes {
		if node.Kind != KindFile {
			continue
		}
		file, err := root.Open(joinNode(name, node.Path))
		if err != nil {
			return Inspection{}, fmt.Errorf("open source %q: %w", joinNode(name, node.Path), err)
		}
		if err := file.Close(); err != nil {
			return Inspection{}, fmt.Errorf("close source %q: %w", joinNode(name, node.Path), err)
		}
	}
	return inspection, nil
}

func copyTree(source *os.Root, sourceName string, destination *os.Root, destinationName string, storageMode bool) error {
	inspection, err := ValidateCopySource(source, sourceName)
	if err != nil {
		return err
	}
	if inspection.Kind == KindAbsent {
		return fmt.Errorf("copy %q: source does not exist", sourceName)
	}
	return copyEntry(source, sourceName, destination, destinationName, storageMode)
}

func Remove(root *os.Root, name string) error {
	if err := root.RemoveAll(name); err != nil {
		return fmt.Errorf("remove %q: %w", name, err)
	}
	return nil
}

func ApplyModes(root *os.Root, base string, nodes []Node) error {
	for _, node := range nodes {
		if node.Kind != KindFile {
			continue
		}
		if err := root.Chmod(joinNode(base, node.Path), node.Mode.Perm()); err != nil {
			return fmt.Errorf("restore file mode %q: %w", joinNode(base, node.Path), err)
		}
	}
	for i := len(nodes) - 1; i >= 0; i-- {
		node := nodes[i]
		if node.Kind != KindDirectory {
			continue
		}
		if err := root.Chmod(joinNode(base, node.Path), node.Mode.Perm()); err != nil {
			return fmt.Errorf("restore directory mode %q: %w", joinNode(base, node.Path), err)
		}
	}
	return nil
}

func copyEntry(source *os.Root, sourceName string, destination *os.Root, destinationName string, storageMode bool) error {
	info, err := source.Lstat(sourceName)
	if err != nil {
		return fmt.Errorf("stat source %q: %w", sourceName, err)
	}
	kind, err := kindOf(info.Mode())
	if err != nil {
		return fmt.Errorf("copy %q: %w", sourceName, err)
	}

	switch kind {
	case KindDirectory:
		createMode := info.Mode().Perm() | 0o700
		finalMode := info.Mode().Perm()
		if storageMode {
			createMode = 0o700
			finalMode = 0o700
		}
		if err := destination.MkdirAll(destinationName, createMode); err != nil {
			return fmt.Errorf("create directory %q: %w", destinationName, err)
		}
		entries, err := fs.ReadDir(source.FS(), sourceName)
		if err != nil {
			return fmt.Errorf("read directory %q: %w", sourceName, err)
		}
		for _, entry := range entries {
			if err := copyEntry(
				source,
				path.Join(sourceName, entry.Name()),
				destination,
				path.Join(destinationName, entry.Name()),
				storageMode,
			); err != nil {
				return err
			}
		}
		if err := destination.Chmod(destinationName, finalMode); err != nil {
			return fmt.Errorf("set directory mode %q: %w", destinationName, err)
		}
	case KindFile:
		if err := destination.MkdirAll(path.Dir(destinationName), 0o755); err != nil {
			return fmt.Errorf("create parent for %q: %w", destinationName, err)
		}
		input, err := source.Open(sourceName)
		if err != nil {
			return fmt.Errorf("open source %q: %w", sourceName, err)
		}
		defer input.Close()
		fileMode := info.Mode().Perm()
		if storageMode {
			fileMode = 0o600
		}
		output, err := destination.OpenFile(destinationName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, fileMode)
		if err != nil {
			return fmt.Errorf("create destination %q: %w", destinationName, err)
		}
		if _, err := io.Copy(output, input); err != nil {
			output.Close()
			return fmt.Errorf("copy %q: %w", sourceName, err)
		}
		if err := output.Close(); err != nil {
			return fmt.Errorf("close destination %q: %w", destinationName, err)
		}
		if err := destination.Chmod(destinationName, fileMode); err != nil {
			return fmt.Errorf("set file mode %q: %w", destinationName, err)
		}
	case KindSymlink:
		if err := destination.MkdirAll(path.Dir(destinationName), 0o755); err != nil {
			return fmt.Errorf("create parent for %q: %w", destinationName, err)
		}
		target, err := source.Readlink(sourceName)
		if err != nil {
			return fmt.Errorf("read symlink %q: %w", sourceName, err)
		}
		if err := destination.Symlink(target, destinationName); err != nil {
			return fmt.Errorf("create symlink %q: %w", destinationName, err)
		}
	}
	return nil
}

func validateParentComponents(root *os.Root, name string) error {
	parts := strings.Split(name, "/")
	for i := 1; i < len(parts); i++ {
		parent := path.Join(parts[:i]...)
		info, err := root.Lstat(parent)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("inspect parent %q: %w", parent, err)
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return fmt.Errorf("managed path %q traverses symlink %q", name, parent)
		}
	}
	return nil
}

func joinNode(base, relative string) string {
	if relative == "." {
		return base
	}
	return path.Join(base, relative)
}

func mustKind(mode fs.FileMode) Kind {
	kind, _ := kindOf(mode)
	return kind
}

func kindOf(mode fs.FileMode) (Kind, error) {
	switch {
	case mode.IsRegular():
		return KindFile, nil
	case mode.IsDir():
		return KindDirectory, nil
	case mode&fs.ModeSymlink != 0:
		return KindSymlink, nil
	default:
		return "", fmt.Errorf("unsupported file type %s", strings.TrimSpace(mode.Type().String()))
	}
}
