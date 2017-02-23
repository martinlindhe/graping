# general TODO

    press V to toggle more display of details

cli flag to toggle --dot (default = braille)
    
braille mode one off innan den börjar scrolla. plus att braille ritas en rad utanför diagrammet


# research termui bugs

linechart-auto-height branch, PULL:
https://github.com/gizak/termui/pull/102
    * LineChart height never resets after it grew, even with different data with lower max pos

linechart-fix-x-label-dynamic-data branch, PULL:
https://github.com/gizak/termui/pull/103
* horiz labels (except 0) are NEVER plotted if initial data len is 0



* if LineChart "braille" mode is used, sometimes it takes 2 ticks before 2 last (braille) symbols is drawn

* top is one off. sometimes a single dot is drawn in Y = 0

    need values to re-create it
