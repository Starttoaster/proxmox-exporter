package proxmox

type BootInfo struct {
	Mode       string `json:"mode"`
	SecureBoot int    `json:"secureboot"`
}

type CpuInfo struct {
	Cores   int    `json:"cores"`
	Cpus    int    `json:"cpus"`
	Flags   string `json:"flags"`
	Hvm     string `json:"hvm"`
	Mhz     string `json:"mhz"`
	Model   string `json:"model"`
	Sockets int    `json:"sockets"`
	UserHz  int    `json:"user_hz"`
}

type CurrentKernel struct {
	Machine string `json:"machine"`
	Release string `json:"release"`
	Sysname string `json:"sysname"`
	Version string `json:"version"`
}

type Ksm struct {
	Shared int `json:"shared"`
}

type Memory struct {
	Free  int64 `json:"free"`
	Total int64 `json:"total"`
	Used  int64 `json:"used"`
}

type RootFs struct {
	Avail int64 `json:"avail"`
	Free  int64 `json:"free"`
	Total int64 `json:"total"`
	Used  int64 `json:"used"`
}

type Swap struct {
	Free  int64 `json:"free"`
	Total int64 `json:"total"`
	Used  int   `json:"used"`
}
