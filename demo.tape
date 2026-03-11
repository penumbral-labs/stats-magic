# Stats Magic — Demo Recording
# Run: vhs demo.tape
# Requires: vhs, ttyd, ffmpeg

Output demo.gif

Set FontSize 14
Set Width 1200
Set Height 750
Set FontFamily "JetBrainsMono Nerd Font"
Set Theme "Catppuccin Mocha"
Set Padding 10
Set Framerate 24
Set TypingSpeed 50ms

# Build silently
Hide
Type "go build -o stats-magic . && clear"
Enter
Sleep 1s
Show

# Launch
Type "./stats-magic"
Enter
Sleep 1s

# Add Fireball
Type "a"
Sleep 300ms
Type "/fire"
Sleep 400ms
Enter
Sleep 300ms
Enter
Sleep 800ms

# Add Lightning Bolt
Type "a"
Sleep 300ms
Type "/light"
Sleep 400ms
Enter 2
Sleep 800ms

# Add Chain Lightning
Type "a"
Sleep 300ms
Type "/chain"
Sleep 400ms
Enter 2
Sleep 800ms

# Add a Fortitude save spell
Type "a"
Sleep 300ms
Type "/fort"
Sleep 400ms
Enter
Sleep 300ms
Type "j"
Enter
Sleep 1s

# Navigate and select for comparison
Type " "
Sleep 300ms
Type "j "
Sleep 300ms
Type "j "
Sleep 300ms
Type "j "
Sleep 1.5s

# Browse the detail panes
Type "kk"
Sleep 800ms
Type "k"
Sleep 1.5s

# Heighten Fireball
Type "++"
Sleep 500ms
Type "++"
Sleep 1.5s

# Edit encounter — change PC level to 10
Tab
Sleep 400ms
Backspace
Type "10"
Sleep 400ms

# Tab to save fields, cycle profiles
Tab 4
Sleep 400ms
Type " "
Sleep 400ms
Type " "
Sleep 400ms
Tab
Type " "
Sleep 1s

# Done editing
Escape
Sleep 1s

# Navigate to show updated stats
Type "j"
Sleep 600ms
Type "j"
Sleep 600ms

# Lower rank back down
Type "--"
Sleep 500ms
Type "--"
Sleep 1.5s

# Final pause
Sleep 2s

Type "q"
Sleep 300ms
