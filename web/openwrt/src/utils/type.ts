export interface Status {
  timestamp: string
  connected: boolean
  mac: string
}

export interface Client {
  ip: string
  mac: string
  phy: string
  hostname: string
  nickName: string
  starTime: string
  online: boolean
  statusList: Status[]
}
