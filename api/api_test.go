package api

import "testing"

func TestDirections(t *testing.T) {
	for _, dir := range []string{"up", "down", "left", "right"} {
		t.Run(dir, func(t *testing.T) {
			c := Coord{
				X: 1,
				Y: 1,
			}
			var next Coord
			switch dir {
			case "up":
				next = c.Up()
			case "down":
				next = c.Down()
			case "left":
				next = c.Left()
			case "right":
				next = c.Right()
			}

			if cn := c.Dir(next); cn != dir {
				t.Logf("c:    %#v", c)
				t.Logf("next: %#v", next)
				t.Errorf("expected to get %s, got: %s", dir, cn)
			}
		})
	}
}
