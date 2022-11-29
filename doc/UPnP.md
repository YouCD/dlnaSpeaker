# 四种服务
* AVTransport Service （可控制多屏设备上的媒体 play，pause，seek，stop 等）
* RenderingControl Service （可调节多屏设备上的音量，声音，静音等）
* ContentDirectory Service （可获取多屏设备上可访问的媒体内容）
* ConnectionManager Service （可提供所支持的传输协议信息及多屏设备的 MIME 格式信息）


```shell


/*
ST：设置服务查询的目标，它必须是下面的类型：
	ssdp:all 搜索所有设备和服务
	upnp:rootdevice 仅搜索网络中的根设备
	uuid:device-UUID 查询UUID标识的设备
	urn:schemas-upnp-org:device:device-
	type:version查询device-Type字段指定的设备类型，设备类型和版本由UPNP组织定义
	urn:schemas-upnp-org:service:service-
	type:version 查询service-Type字段指定的服务类型，服务类型和版本由UPNP组织定义
*/

/* UPnP 搜索 报文
   M-SEARCH * HTTP/1.1           // 固定
   ST: upnp:rootdevice			 // 指定了特殊类型设备 也可以是  ssdp:all [表示 搜索所有]
   MX: 3						 // 等待时长
   MAN: "ssdp:discover" 		 // 固定
   HOST: 239.255.255.250:1900    // 多播地址
   ===
   HOST: 239.255.255.250:1900
   NTS: ssdp:alive
   USN: uuid:ab69a6f5-c04f-4d12-ba58-4aa11381da13::urn:schemas-upnp-org:service:AVTransport:1
   LOCATION: http://192.168.1.13:4564/description.xml
   EXT:
   SERVER: Linux/6.0.9-arch1-1 UPnP/1.0 Macast/0.7
   CACHE-CONTROL: max-age=66
   NT: urn:schemas-upnp-org:service:AVTransport:1

 返回设备信息

	NOTIFY * HTTP/1.1
	HOST: 239.255.255.250:1900
	NTS: ssdp:alive
	USN: uuid:a11d2b86-60c2-42f4-9ab7-cd6c752d59ec::urn:schemas-upnp-org:device:MediaRenderer:1
	LOCATION: http://192.168.1.13:4564/description.xml
	EXT:
	SERVER: Linux/6.0.7-arch1-1 UPnP/1.0 Macast/0.7
	CACHE-CONTROL: max-age=66
	NT: urn:schemas-upnp-org:device:MediaRenderer:1


*/
```