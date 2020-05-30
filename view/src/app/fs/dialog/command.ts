export enum NetCommand {
    // 錯誤
    Error = 1,
    // websocket 心跳防止瀏覽器 關閉不獲取 websocket
    Heart = 2,
    // 更新進度
    Progress = 3,
    // 操作完成
    Done = 4,
    // 初始化
    Init = 5,
    //  確認操作
    Yes = 6,
    // 取消操作
    No = 7,
    // 檔案已經存在
    Exist = 8,
    // 覆蓋全部 重複檔案
    YesAll = 9,
    // 跳過 重複檔案
    Skip = 10,
    // 跳過全部 重複檔案
    SkipAll = 11,
}
