package utils

func IsValidLuhnNumber(number int) bool {
	return (number%10+checkSumLuhnNumber(number/10))%10 == 0
}

func checkSumLuhnNumber(number int) int {
	var luhn int

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 {
			cur *= 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number /= 10
	}

	return luhn % 10
}
