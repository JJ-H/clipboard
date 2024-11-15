// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {main} from '../models';

export function AddTag(arg1:string,arg2:string):Promise<void>;

export function DeleteHistoryItem(arg1:string):Promise<void>;

export function DeleteTag(arg1:string):Promise<void>;

export function GetConfig():Promise<main.Config>;

export function GetHistory():Promise<Array<main.ClipboardItem>>;

export function HideWindow():Promise<void>;

export function MinimizeWindow():Promise<void>;

export function MoveItemToFront(arg1:string):Promise<void>;

export function QuitApp():Promise<void>;

export function SaveToClipboard(arg1:string):Promise<void>;

export function ShowWindow():Promise<void>;

export function ToggleWindow():Promise<void>;

export function UpdateConfig(arg1:number,arg2:boolean):Promise<void>;

export function UpdateItemTag(arg1:string,arg2:string):Promise<void>;

export function UpdateTag(arg1:string,arg2:string,arg3:string):Promise<void>;

export function UpdateTagsOrder(arg1:Array<string>):Promise<void>;
