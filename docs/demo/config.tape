Output docs/assets/config.gif
Require homelabctl
Set Shell "zsh"
Set Width 1100
Set Height 800
Set FontSize 20
Set TypingSpeed 40ms
Set Framerate 15
Set Theme "Kanagawa"

# Reset only the generated fixture inventory; visible commands use defaults.
Hide
Type `export XDG_CONFIG_HOME="$PWD/bin/form-demo"`
Enter
Type "mkdir -p bin/form-demo/homelabctl && cp docs/demo/tui.toml bin/form-demo/homelabctl/config.toml"
Enter
Sleep 500ms
Type "export PS1='$ '; clear"
Enter
Sleep 500ms
Show

Type "homelabctl config add uptime"
Enter
Sleep 1s
Type "uptime.example.com"
Tab
Sleep 500ms
Type "192.0.2.10"
Tab
Sleep 500ms
Type "443"
Tab
Sleep 1s

# Keep HTTP and TCP; deselect DNS and TLS.
Down
Sleep 500ms
Space
Down 2
Sleep 500ms
Space
Sleep 1s
Tab
Enter
Sleep 1s
Tab
Sleep 500ms
Type "uptime.md"
Tab
Sleep 500ms
Type "200"
Tab
Sleep 500ms
Type "3s"
Tab
Sleep 2s
Enter
Sleep 1s

# Reopen the saved service, update its port and timeout, then save again.
Type "homelabctl config edit uptime"
Sleep 500ms
Enter
Sleep 1s
Tab 2
Sleep 500ms
Ctrl+U
Sleep 200ms
Type "8443"
Sleep 200ms
Tab 2
Sleep 500ms
Enter
Tab 3
Sleep 500ms
Ctrl+U
Sleep 200ms
Type "5s"
Tab
Sleep 2s
Enter
Sleep 1s
Type "homelabctl list"
Enter
Sleep 3s
