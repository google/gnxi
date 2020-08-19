import { Component, OnInit, SimpleChanges, Input, OnChanges } from '@angular/core';
import { Devices, GetDevices, Device } from '../../api/devices'
import { FormControl } from '@angular/forms';

@Component({
  selector: 'app-devices',
  templateUrl: './devices.component.html',
  styleUrls: ['./devices.component.css']
})
export class DevicesComponent implements OnInit {

  constructor() { }

  ngOnInit(): void {
    this.selectedDeviceName.registerOnChange(() => {
      console.log(this.selectedDeviceName.value);
    })
  }
  deviceList: Devices = GetDevices();

  deviceNameList: string[] = Object.keys(this.deviceList);
  selectedDeviceName = new FormControl("");

  caCert = new FormControl("");
}
