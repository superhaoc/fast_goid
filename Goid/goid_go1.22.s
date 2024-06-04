#include "textflag.h"

TEXT ·getg(SB),NOSPLIT,$0-8
#ifdef GOARCH_amd64
	MOVQ (TLS), AX
	MOVQ AX, ret+0(FP)
#else
    MOVQ $0, ret+0(FP)
#endif
    RET
