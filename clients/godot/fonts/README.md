# Fonts

## Mx437_IBM_VGA_9x16.ttf

Cell-tight reproduction of the classic DOS VGA 9×16 text-mode font. Used by
the terminal panel so stacked CP437 block glyphs (`█▄▀▌▐`) tile seamlessly
and line descenders aren't clipped.

Source: **VileR / int10h.org** — "Oldschool PC Font Resource"
<https://int10h.org/oldschool-pc-fonts/>

License: **CC-BY-SA 4.0** (<https://creativecommons.org/licenses/by-sa/4.0/>)

The `Mx` variant ("mixed outline + bitmap") embeds the 9×16 bitmap at native
size and provides TrueType outlines for any other size — robust for Godot's
scaled rendering.
