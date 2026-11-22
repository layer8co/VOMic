package ir.layer8.vomic

import java.nio.ByteBuffer
import java.nio.ByteOrder

// Helper types, useful for network operations.
data class SessionInit(
    val flag: UByte,
    val hostname: String,
)

data class SessionAck(
    val media_port: UShort
)

fun SessionInit.marshalBinary(): ByteBuffer {
    val hostnameBytes = this.hostname.toByteArray(Charsets.UTF_8)
    val cap = UByte.SIZE_BYTES + hostnameBytes.size + 1
    val buffer = ByteBuffer.allocate(cap)
    buffer.order(ByteOrder.BIG_ENDIAN)

    buffer.put(flag.toByte())
    buffer.put(hostnameBytes)
    buffer.put(0)
    buffer.flip()

    return buffer
}

fun ByteBuffer.toSessionAck(): SessionAck {
    this.order(ByteOrder.BIG_ENDIAN)
    this.position(0)
    val port = this.getShort().toUShort()

    return SessionAck(port)
}
