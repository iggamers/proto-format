package proto_format

import (
	"bytes"
	"fmt"
	"github.com/emicklei/proto"
	"github.com/just-bytes/proto-format/pkg"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

func init() {
	var getServeDir = func(path string) string {
		var run, _ = os.Getwd()
		return strings.Replace(path, run, ".", -1)
	}
	var formatter = &nested.Formatter{
		NoColors:        false,
		HideKeys:        true,
		TimestampFormat: "2006-01-02 15:04:05",
		CallerFirst:     true,
		CustomCallerFormatter: func(f *runtime.Frame) string {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			return fmt.Sprintf(" [%s:%d][%s()]", getServeDir(f.File), f.Line, funcName)
		},
	}
	if runtime.GOOS == "windows" {
		formatter.NoColors = true
	}
	log.SetFormatter(formatter)
	log.SetReportCaller(true)
}

func listProtoFile(path string) []string {
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	var result []string
	fi, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, file := range fi {
		if file.IsDir() {
			result = append(result, listProtoFile(path+"/"+file.Name())...)
		}
		if strings.HasSuffix(file.Name(), ".proto") {
			result = append(result, path+"/"+file.Name())
		}
	}
	return result
}

// Format 格式化 简单的目录、通配符支持 遍历有问题 须修复
func Format(pbFile string) error {
	var pbFileList []string
	fi, err := os.Stat(pbFile)
	if err == nil && fi.IsDir() {
		// 是目录 需要遍历目录 扫描一下 .proto结尾的文件
		pbFileList = listProtoFile(pbFile)
	} else if strings.HasSuffix(pbFile, "*.proto") || strings.HasSuffix(pbFile, "*") {
		pbFileList = listProtoFile(path.Dir(pbFile))
	}

	defer func() {
		if len(pbFile) != 0 {
			_, _ = fmt.Fprintf(os.Stdout, "格式化列表: %+v\n", pbFileList)
		}
	}()

	// 取目录 取文件名
	dirName := path.Dir(pbFile)
	baseName := path.Base(pbFile)

	if strings.HasSuffix(baseName, ".proto") {
		if err := readFormatWrite(dirName + "/" + baseName); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "格式化失败: %+v\n", err)
			return err
		}
		return nil
	}

	for _, pbPath := range pbFileList {
		if err := readFormatWrite(pbPath); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "格式化失败: %+v\n", err)
			return err
		}
	}

	return nil
}

// readFormatWrite 写入
func readFormatWrite(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "格式化失败: %+v\n", err)
		return err
	}
	defer file.Close()
	buf := new(bytes.Buffer)
	if err := format(filename, file, buf); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "格式化失败: %+v\n", err)
		return err
	}

	if err := ioutil.WriteFile(filename, buf.Bytes(), os.ModePerm); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "格式化失败: %+v\n", err)
		return err
	}

	return nil
}

func format(filename string, input io.Reader, output io.Writer) error {
	parser := proto.NewParser(input)
	parser.Filename(filename)
	def, err := parser.Parse()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "格式化失败: %+v\n", err)
		return err
	}
	pkg.NewFormatter(output, "  ").Format(def) // 4 spaces
	return nil
}
