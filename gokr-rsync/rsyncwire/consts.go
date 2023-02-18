package rsyncwire

// ProtocolVersion defines the currently implemented rsync protocol
// version. Protocol version 27 seems to be the safest bet for wide
// compatibility: version 27 was introduced by rsync 2.6.0 (released 2004), and
// is supported by openrsync and rsyn.
const ProtocolVersion = 27
