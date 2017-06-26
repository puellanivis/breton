// func bsr(u uint64) uint64
TEXT ·bsr(SB),$0
	MOVQ u+0(FP), AX
	BSRQ AX, AX
	JNZ ret
	XORQ AX, AX
ret:
	MOVQ AX, r+8(FP)
	RET
