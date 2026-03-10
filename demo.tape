# Stats Magic — Demo Recording
# Run: vhs demo.tape
# Requires: vhs, ttyd, ffmpeg

Output demo.gif

Set FontSize 13
Set Width 1400
Set Height 900
Set FontFamily "JetBrains Mono"
Set Theme "Catppuccin Mocha"
Set Padding 10
Set Framerate 24
Set TypingSpeed 75ms

Require stats-magic

# Build silently
Hide
Type "go build -o stats-magic . && clear"
Enter
Sleep 1s
Show

# Launch
Type "./stats-magic"
Enter
Sleep 1.5s

# Add Fireball from presets
Type "a"
Sleep 500ms
Type "fire"
Sleep 500ms
Enter
Sleep 1s

# Add another spell — Electric Arc
Type "a"
Sleep 500ms
Type "electric"
Sleep 500ms
Enter
Sleep 1s

# Add Heal (Harm variant)
Type "a"
Sleep 500ms
Type "harm"
Sleep 500ms
Enter
Sleep 1s

# Navigate the list
Down
Sleep 400ms
Down
Sleep 400ms
Up
Sleep 400ms
Up
Sleep 600ms

# Heighten Fireball
Type "+"
Sleep 500ms
Type "+"
Sleep 500ms
Type "+"
Sleep 800ms

# Select spells for comparison
Space
Sleep 300ms
Down
Sleep 300ms
Space
Sleep 300ms
Down
Sleep 300ms
Space
Sleep 1s

# Edit encounter parameters
Tab
Sleep 500ms

# Change PC level
Backspace
Backspace
Type "10"
Tab
Sleep 500ms
Tab
Sleep 300ms
Tab
Sleep 300ms

# Cycle a save profile
Space
Sleep 400ms
Space
Sleep 400ms

# Done editing encounter
Escape
Sleep 1.5s

# Navigate back through spells to show detail
Up
Sleep 600ms
Up
Sleep 600ms

# Lower rank back down
Type "-"
Sleep 400ms
Type "-"
Sleep 600ms

# Pause to admire
Sleep 2s

# Quit
Type "q"
Sleep 500ms
