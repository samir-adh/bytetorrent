# Notes

now that we are being unchoked by peers we can start requesting for pieces which rises a few problems to tackle

- because each peer have a different set of pieces we need to track which pieces are available across different coroutines
- we need to handle what happens when we downloaded all pieces needed from a peer
- we need to handle what happens when a peer stop responding

the ideal download process would be

TorrentClient spawns PeerConnection
PeerConnection connect to peer and returns the bitfield of length N
TC creates a DS "availablePieces" a list that for each piece record its index, if it is downloaded and the PC instances that can get it and if they are available
Then for each piece it sends the next available PC to get it if no PC is available skip to the next piece
Once a PC has finished it returns it to the TC which will write it on the disk


for now our most important problem is to download a piece from a peer, block by block so let's start with this.
