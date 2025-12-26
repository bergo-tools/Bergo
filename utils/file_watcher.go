package utils

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher 监听工作目录下的文件改动
// 当文件包含 @bergo 标记时，通知回调
type FileWatcher struct {
	watcher    *fsnotify.Watcher
	workDir    string
	callback   func(filePath string)
	stopCh     chan struct{}
	mu         sync.Mutex
	running    bool
	checkedMap map[string]bool // 记录已检查过的文件，避免重复通知
}

// NewFileWatcher 创建新的文件监听器
func NewFileWatcher(workDir string, callback func(filePath string)) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw := &FileWatcher{
		watcher:    watcher,
		workDir:    workDir,
		callback:   callback,
		stopCh:     make(chan struct{}),
		checkedMap: make(map[string]bool),
	}

	return fw, nil
}

// Start 开始监听
func (fw *FileWatcher) Start() error {
	fw.mu.Lock()
	if fw.running {
		fw.mu.Unlock()
		return nil
	}
	fw.running = true
	fw.checkedMap = make(map[string]bool) // 重置已检查的文件
	fw.mu.Unlock()

	// 添加工作目录到监听
	if err := fw.addWatchRecursive(fw.workDir); err != nil {
		return err
	}

	// 启动监听协程
	go fw.watchLoop()

	return nil
}

// Stop 停止监听
func (fw *FileWatcher) Stop() {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if !fw.running {
		return
	}
	fw.running = false
	close(fw.stopCh)
	fw.watcher.Close()
}

// watchLoop 监听循环
func (fw *FileWatcher) watchLoop() {
	for {
		select {
		case <-fw.stopCh:
			return
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			fw.handleEvent(event)
		case _, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			// 忽略错误，继续监听
		}
	}
}

// handleEvent 处理文件事件
func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
	// 只处理写入和创建事件
	if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
		return
	}

	filePath := event.Name

	// 检查是否应该忽略该文件
	if fw.shouldIgnore(filePath) {
		return
	}

	// 检查文件是否包含 @bergo 标记
	if fw.containsBergoTag(filePath) {
		fw.mu.Lock()
		// 避免重复通知同一个文件
		if !fw.checkedMap[filePath] {
			fw.checkedMap[filePath] = true
			fw.mu.Unlock()
			if fw.callback != nil {
				fw.callback(filePath)
			}
		} else {
			fw.mu.Unlock()
		}
	}
}

// shouldIgnore 检查是否应该忽略该文件
func (fw *FileWatcher) shouldIgnore(filePath string) bool {
	// 获取文件名
	fileName := filepath.Base(filePath)

	// 忽略隐藏文件（以.开头）
	if strings.HasPrefix(fileName, ".") {
		return true
	}

	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if err != nil {
		return true
	}

	// 忽略目录
	if info.IsDir() {
		return true
	}

	// 忽略二进制文件
	isBinary, err := IsBinaryFile(filePath)
	if err != nil || isBinary {
		return true
	}

	return false
}

// containsBergoTag 检查文件是否包含 @bergo 标记
func (fw *FileWatcher) containsBergoTag(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	return strings.Contains(string(content), "@bergo")
}

// addWatchRecursive 递归添加目录到监听
func (fw *FileWatcher) addWatchRecursive(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续遍历
		}

		// 只监听目录
		if !info.IsDir() {
			return nil
		}

		// 忽略隐藏目录
		if strings.HasPrefix(info.Name(), ".") && path != dir {
			return filepath.SkipDir
		}

		// 忽略常见的不需要监听的目录
		ignoreDirs := []string{"node_modules", "vendor", ".git", "__pycache__", "dist", "build"}
		for _, ignoreDir := range ignoreDirs {
			if info.Name() == ignoreDir {
				return filepath.SkipDir
			}
		}

		fw.watcher.Add(path)
		return nil
	})
}
