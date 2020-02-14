package command

type Server struct {
	ServerName     string `toml:"server_name"`
	Hostname       string `toml:"hostname"`
	Port           int    `toml:"port"`
	User           string `toml:"user"`
	Password       string `toml:"password"`
	PrivateKeyPath string `toml:"private_key_path"`
	TailFile       string `toml:"tail_file"`
	TailLine       int
}
