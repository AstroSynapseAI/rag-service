package documents

type DocuemntsToolOption func(*DocumentsTool)

func WithRootPath(path string) DocuemntsToolOption {
	return func(tool *DocumentsTool) {
		tool.RootPath = path
	}
}