package util

import "github.com/toujourser/gomoku/internal/entity"

func HasStep(i int8, j int8, color int8, steps *[]entity.Chess) bool {
	for k := int(color); k < len(*steps); k += 2 {
		step := (*steps)[k]
		if step.I == i && step.J == j {
			return true
		}
	}
	return false
}

// checkFiveInDirection 检查各个方向是否达成五子连珠
// 接受起始坐标 (i, j)、棋子颜色 color、在 x 和 y 方向上的增量 x 和 y，以及棋局步骤切片的指针 steps。
// 它在给定的方向上遍历棋局，计算连续相同颜色的棋子数量。
// 如果连续棋子数量达到 5 或更多，则返回 true，表示达成了五子连珠的条件。
func checkFiveInDirection(i int8, j int8, color int8, x int8, y int8, steps *[]entity.Chess) bool {
	count := 1
	for m, n := i-x, j-y; m >= 0 && n >= 0 && m < 15 && n < 15; m, n = m-x, n-y {
		if HasStep(m, n, color, steps) {
			count++
		} else {
			break
		}
	}
	for m, n := i+x, j+y; m >= 0 && n >= 0 && m < 15 && n < 15; m, n = m+x, n+y {
		if HasStep(m, n, color, steps) {
			count++
		} else {
			break
		}
	}

	return count >= 5
}

// CheckFiveOfLastStep 检查最后一步棋是否导致了五子连珠
// 函数接受一个棋局步骤的切片 steps 的指针作为参数。
// 首先，它确定最后一步棋的颜色（根据当前步骤数的奇偶性来确定）。
// 如果步骤数小于 9，表示棋局还没有达到五子连珠的可能性，因此返回 false 和颜色。
func CheckFiveOfLastStep(steps *[]entity.Chess) (bool, int8) {
	color := int8((len(*steps) - 1) % 2)
	if len(*steps) < 9 {
		return false, color
	}
	lastStep := (*steps)[len(*steps)-1]
	i := lastStep.I
	j := lastStep.J

	// 函数获取最后一步棋的坐标，并在四个方向上检查是否有五子连珠的情况。
	// 调用 checkFiveInDirection() 函数四次，分别检查水平、垂直、正斜线和反斜线方向上的连珠情况。
	// 如果任意一个方向上有五子连珠，则返回 true 和颜色。
	hasFive := checkFiveInDirection(i, j, color, 1, 0,
		steps) || checkFiveInDirection(i, j, color, 0, 1,
		steps) || checkFiveInDirection(i, j, color, 1, 1,
		steps) || checkFiveInDirection(i, j, color, 1, -1, steps)

	return hasFive, color
}
