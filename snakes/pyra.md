# pyra

A goal-oriented snake AI.

## Goals

- Gain a minimum snake length
  - 8
  - 9
  - 10
- Chase tail
- Search food at 30 health or lower
- Attack the heads of enemy snakes

## Targeting

Possible targets:

- food
  - if health <= 30:
    - score = 9000
  - if len(me) < 8:
    - score = 50
  - score = 20
- own tail
  - score = 30
- enemy snake head
  - if len(me) > len(enemy):
    - score = 400
- subtract the manhattan distance from the score

for every target:

- attempt to find a path to the target via a-star
- save the astar length
- return the target with the highest score and lowest length
