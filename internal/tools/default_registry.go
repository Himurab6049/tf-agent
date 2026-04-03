package tools

// NewDefaultRegistry creates a registry with all built-in tools registered.
// cwd is the working directory used by filesystem tools (read, write, edit, glob, grep, ls).
func NewDefaultRegistry(cwd string) *Registry {
	r := NewRegistry()
	r.Register(&BashTool{})
	r.Register(NewReadTool(cwd))
	r.Register(NewWriteTool(cwd))
	r.Register(NewEditTool(cwd))
	r.Register(NewGlobTool(cwd))
	r.Register(NewGrepTool(cwd))
	r.Register(NewLsTool(cwd))
	r.Register(NewWebFetchTool())
	r.Register(NewWebSearchTool())
	r.Register(&TaskTool{})
	return r
}
