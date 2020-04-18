package config

type UploaderCheck struct {
	Type    string
	Limit   uint64
	Exclude []string
	Include []string
}

type UploaderHidden struct {
	Enabled bool
	Type    string
	Folder  string
	Cleanup bool
	Workers int
}

type UploaderRemotesMoveServerSide struct {
	From string
	To   string
}

type UploaderRemotes struct {
	Clean          []string
	Copy           []string
	Move           string
	MoveServerSide []UploaderRemotesMoveServerSide `mapstructure:"move_server_side"`
	Dedupe         []string
}

type UploaderRcloneParams struct {
	Copy           []string
	Move           []string
	MoveServerSide []string `mapstructure:"move_server_side"`
}

type UploaderConfig struct {
	Enabled              bool
	Check                UploaderCheck
	Hidden               UploaderHidden
	LocalFolder          string `mapstructure:"local_folder"`
	ServiceAccountFolder string `mapstructure:"sa_folder"`
	Remotes              UploaderRemotes
	RcloneParams         UploaderRcloneParams `mapstructure:"rclone_params"`
}
