from scapy.all import *
 
protocols = {1:'ICMP', 6:'TCP', 17:'UDP'}
 
def showPacket(packet):
    src_ip = packet[0][1].src
    dst_ip = packet[0][1].dst
    proto = packet[0][1].proto

    if proto in protocols:
        print ("%s %s !" %(packet[0][0].src, packet[0][0].dst))
        print ("protocol: %s: %s -> %s" %(protocols[proto], src_ip, dst_ip))
 
        if proto == 6 or proto == 17:
            print( packet[0][2].chksum)
            del packet[0][2].chksum
            recom = IP(str(packet[0]))
            print(recom[2].chksum)
            print("%d: %s" %(len(packet[0][2].payload), packet[0][2].payload))

        if proto == 1:
            print("TYPE: [%d], CODE[%d]" %(packet[0][2].type, packet[0][2].code))

def sniffing(filter):
    sniff(filter = filter, prn = showPacket, count = 0)
 
if __name__ == '__main__':
    filter = 'ip'
    sniffing(filter)
