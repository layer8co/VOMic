package ir.layer8.vomic

import org.junit.Test
import org.junit.Assert.*
import java.nio.ByteBuffer

class ProtocolTest {

    @Test
    fun `session init produces correct binary representation`() {
        val session = SessionInit(
            flag = 1u,
            hostname = "Xiaomi Redmi note"
        )

        val buffer = session.marshalBinary()

        val expected = byteArrayOf(
            1,
            *session.hostname.toByteArray(),
            0,
        )

        val actual = ByteArray(buffer.remaining())
        buffer.get(actual)

        assertArrayEquals(expected, actual)
    }


    @Test
    fun `ByteBuffer should produce correct SessionAck representation`() {
        val media_port: UShort = 8125u
        val buffer = ByteBuffer.allocate(1024)
        buffer.putShort(media_port.toShort())

        val sessionAck = buffer.toSessionAck()

        assertEquals(media_port, sessionAck.media_port)
    }
}
