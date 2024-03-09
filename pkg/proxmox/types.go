package proxmox

// BootInfo info about host boot
type BootInfo struct {
	Mode       string `json:"mode"`
	SecureBoot int    `json:"secureboot"`
}

// CPUInfo info about host CPU
type CPUInfo struct {
	Cores   int    `json:"cores"`
	Cpus    int    `json:"cpus"`
	Flags   string `json:"flags"`
	Hvm     string `json:"hvm"`
	Mhz     string `json:"mhz"`
	Model   string `json:"model"`
	Sockets int    `json:"sockets"`
	UserHz  int    `json:"user_hz"`
}

// CurrentKernel info about host kernel
type CurrentKernel struct {
	Machine string `json:"machine"`
	Release string `json:"release"`
	Sysname string `json:"sysname"`
	Version string `json:"version"`
}

// Ksm info about Kernel same-page merging
type Ksm struct {
	Shared int `json:"shared"`
}

// Memory info about host memory
type Memory struct {
	Free  int `json:"free"`
	Total int `json:"total"`
	Used  int `json:"used"`
}

// RootFs info about the host root filesystem
type RootFs struct {
	Avail int `json:"avail"`
	Free  int `json:"free"`
	Total int `json:"total"`
	Used  int `json:"used"`
}

// Swap info about swap
type Swap struct {
	Free  int `json:"free"`
	Total int `json:"total"`
	Used  int `json:"used"`
}
