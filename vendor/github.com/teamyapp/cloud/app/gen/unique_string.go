package gen

type UniqueString struct {
	uniqueNumGen *UniqueNumber
	stringLen    int
	alphabet     []rune
}

func (u UniqueString) GenerateUniqueString() (string, error) {
	currNum, err := u.uniqueNumGen.GenerateUniqueNumber()
	if err != nil {
		return "", err
	}

	return u.toString(currNum), nil
}

func (u UniqueString) toString(num uint64) string {
	base := uint64(len(u.alphabet))
	resultRunes := make([]rune, u.stringLen)
	for strRuneIndex := 0; strRuneIndex < u.stringLen; strRuneIndex++ {
		alphabetRuneIndex := num % base
		num /= base
		alphabetRune := u.alphabet[alphabetRuneIndex]
		resultRunes[strRuneIndex] = alphabetRune
	}

	return string(resultRunes)
}

func NewUniqueString(
	name string,
	stringLen int,
	alphabet string,
	uniqueNumFactory UniqueNumberFactory,
) (UniqueString, error) {
	numNum, err := uniqueNumFactory.MakeUniqueNumber(name)
	if err != nil {
		return UniqueString{}, err
	}

	return UniqueString{
		uniqueNumGen: numNum,
		stringLen:    stringLen,
		alphabet:     []rune(alphabet),
	}, nil
}
