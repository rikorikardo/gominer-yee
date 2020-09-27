package yee

const kernelSource = `__constant static const uchar blake2b_sigma[12][16] = {
	{ 0,  1,  2,  3,  4,  5,  6,  7,  8,  9,  10, 11, 12, 13, 14, 15 } ,
	{ 14, 10, 4,  8,  9,  15, 13, 6,  1,  12, 0,  2,  11, 7,  5,  3  } ,
	{ 11, 8,  12, 0,  5,  2,  15, 13, 10, 14, 3,  6,  7,  1,  9,  4  } ,
	{ 7,  9,  3,  1,  13, 12, 11, 14, 2,  6,  5,  10, 4,  0,  15, 8  } ,
	{ 9,  0,  5,  7,  2,  4,  10, 15, 14, 1,  11, 12, 6,  8,  3,  13 } ,
	{ 2,  12, 6,  10, 0,  11, 8,  3,  4,  13, 7,  5,  15, 14, 1,  9  } ,
	{ 12, 5,  1,  15, 14, 13, 4,  10, 0,  7,  6,  3,  9,  2,  8,  11 } ,
	{ 13, 11, 7,  14, 12, 1,  3,  9,  5,  0,  15, 4,  8,  6,  2,  10 } ,
	{ 6,  15, 14, 9,  11, 3,  0,  8,  12, 2,  13, 7,  1,  4,  10, 5  } ,
	{ 10, 2,  8,  4,  7,  6,  1,  5,  15, 11, 9,  14, 3,  12, 13, 0  } ,
	{ 0,  1,  2,  3,  4,  5,  6,  7,  8,  9,  10, 11, 12, 13, 14, 15 } ,
	{ 14, 10, 4,  8,  9,  15, 13, 6,  1,  12, 0,  2,  11, 7,  5,  3  } };
	
	#define ROTL64( x, n ) ROTR64( x, ( 64 - ( n ) ) )
	#define ROTR64( x, n ) ( ( n ) < 32 ? ( amd_bitalign( (uint)( ( x ) >> 32 ), (uint)( x ), (uint)( n ) ) | ( (ulong)amd_bitalign( (uint)( x ), (uint)( ( x ) >> 32 ), (uint)( n ) ) << 32 ) )\
							: ( amd_bitalign( (uint)( x ), (uint)( ( x ) >> 32 ), (uint)( n ) - 32 ) | ( (ulong)amd_bitalign( (uint)( ( x ) >> 32 ), (uint)( x ), (uint)( n ) - 32 ) << 32 ) ) )
	
#define G_BLAKE2b( r, i, a, b, c, d ) \
{\
	a = a + b + m[ blake2b_sigma[ r ][ 2 * i + 0 ] ];\
	d = as_ulong( as_uint2( d ^ a ).s10 );\
	c = c + d;\
	b = ROTR64( b ^ c, 24 );\
	a = a + b + m[ blake2b_sigma[ r ][ 2 * i + 1 ] ];\
	d = ROTR64( d ^ a, 16 );\
	c = c + d;\
	b = ROTR64( b ^ c, 63 );\
}

#define ROUND_BLAKE2b(r) \
{\
	G_BLAKE2b( r, 0, v[ 0 ], v[ 4 ], v[ 8 ], v[ 12 ] );\
	G_BLAKE2b( r, 1, v[ 1 ], v[ 5 ], v[ 9 ], v[ 13 ] );\
	G_BLAKE2b( r, 2, v[ 2 ], v[ 6 ], v[ 10 ], v[ 14 ] );\
	G_BLAKE2b( r, 3, v[ 3 ], v[ 7 ], v[ 11 ], v[ 15 ] );\
	G_BLAKE2b( r, 4, v[ 0 ], v[ 5 ], v[ 10 ], v[ 15 ] );\
	G_BLAKE2b( r, 5, v[ 1 ], v[ 6 ], v[ 11 ], v[ 12 ] );\
	G_BLAKE2b( r, 6, v[ 2 ], v[ 7 ], v[ 8 ], v[ 13 ] );\
	G_BLAKE2b( r, 7, v[ 3 ], v[ 4 ], v[ 9 ], v[ 14 ] );\
}

#define BLAKE2b_ROUNDS \
	ROUND_BLAKE2b( 0 );\
	ROUND_BLAKE2b( 1 );\
	ROUND_BLAKE2b( 2 );\
	ROUND_BLAKE2b( 3 );\
	ROUND_BLAKE2b( 4 );\
	ROUND_BLAKE2b( 5 );\
	ROUND_BLAKE2b( 6 );\
	ROUND_BLAKE2b( 7 );\
	ROUND_BLAKE2b( 8 );\
	ROUND_BLAKE2b( 9 );\
	ROUND_BLAKE2b( 0 );\
	ROUND_BLAKE2b( 1 );
	
// Target is passed in via headerIn[32 - 29]
__kernel void nonceGrind(__global ulong *headerIn, __global ulong *nonceOut) {
	ulong target = headerIn[4];
	ulong m[16] = {	headerIn[0], headerIn[1],
	                headerIn[2], headerIn[3],
	                (ulong)get_global_id(0), headerIn[5],
	                headerIn[6], headerIn[7],
	                headerIn[8], headerIn[9], 0, 0, 0, 0, 0, 0 };
	ulong v[16] = { 0x6a09e667f2bdc928, 0xbb67ae8584caa73b, 0x3c6ef372fe94f82b, 0xa54ff53a5f1d36f1,
	                0x510e527fade682d1, 0x9b05688c2b3e6c1f, 0x1f83d9abfb41bd6b, 0x5be0cd19137e2179,
	                0x6a09e667f3bcc908, 0xbb67ae8584caa73b, 0x3c6ef372fe94f82b, 0xa54ff53a5f1d36f1,
	                0x510e527fade68281, 0x9b05688c2b3e6c1f, 0xe07c265404be4294, 0x5be0cd19137e2179 };

BLAKE2b_ROUNDS;
									
	if (as_ulong(as_uchar8(0x6a09e667f2bdc928 ^ v[0] ^ v[8]).s76543210) <= target) {
		*nonceOut = m[4];
		return;
	}
}
`
