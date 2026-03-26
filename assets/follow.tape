# Where should we write the GIF?
Output follow.gif

# Set up a 1200x600 terminal with 46px font.
Set Shell bash
Set FontSize 18
Set FontFamily "JetBrainsMono Nerd Font Mono"
Set Width 1200
Set Height 600
Set Padding 10

# Type a command in the terminal.
Type "peacock -f production.log"

# Pause for dramatic effect...
Sleep 1s

# Run the command by pressing enter.
Enter

# Admire the output for a bit.
Sleep 10s

Type "/"

Sleep 500ms

Type@500ms "error"

Sleep 10s

Ctrl+C

Sleep 1s

