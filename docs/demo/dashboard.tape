# Run from the repository root after building and starting the demo server.
Output docs/assets/dashboard.gif
Require homelabctl

Set Shell "zsh"
Set Width 1300
Set Height 600
Set FontSize 20
Set TypingSpeed 35ms
Set Framerate 15
Set CursorBlink false
Set Theme "Kanagawa"

# Hide shell setup so personal prompts do not appear in the recording.
Hide
Type "export PS1='$ '; clear"
Enter
Sleep 500ms
Show

Type "homelabctl dashboard"
Sleep 700ms
Enter
Sleep 3s

# Navigate to a healthy service and inspect its details.
Down
Sleep 1s
Enter
Sleep 2s
Escape
Sleep 500ms

# The offline service shows the actual HTTP and TCP connection errors.
Up
Sleep 500ms
Enter
Sleep 3s
Escape

Sleep 1s
Type "q"
Sleep 500ms
