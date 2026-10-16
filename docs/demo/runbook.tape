Output docs/assets/runbook.gif
Require homelabctl
Require vim
Set Shell "zsh"
Set Width 1100
Set Height 720
Set FontSize 20
Set TypingSpeed 35ms
Set Framerate 15
Env EDITOR "vim"
# A predictable editor session without personal Vim settings or swap files.
Env VIMINIT "set nocompatible noswapfile nobackup nowritebackup viminfo= | syntax on"

Hide
Type `export XDG_CONFIG_HOME="$PWD/bin/runbook-demo"`
Enter
Type "mkdir -p bin/runbook-demo/homelabctl && cp docs/demo/tui.toml bin/runbook-demo/homelabctl/config.toml && cp docs/demo/grafana.md bin/runbook-demo/homelabctl/grafana.md"
Enter
Sleep 500ms
Type "export PS1='$ '; clear"
Enter
Sleep 500ms
Show

Type "homelabctl runbook --style tokyo-night --width 80 grafana"
Enter
Sleep 3s
Down@150ms 10
Sleep 2s

# Open the configured editor and update a note in the generated copy.
Type "e"
Sleep 2s
Type "/Resolution: pending."
Enter
Sleep 600ms
Type "cc- Resolution: corrected the proxy upstream port."
Escape
Sleep 2s
Type ":wq"
Enter
Sleep 2s

# Scroll to the end to show that the viewer reloaded the saved Markdown.
Down@80ms 30
Sleep 1s
Type "q"
Sleep 500ms
