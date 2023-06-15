var lastReceived = 0;
var isWait = false;   //是否等待服务端的响应

var fetch = function () {
    if (isWait) return;     //如果上一次的请求没有得到响应，就一直保持长连接，直到收到服务端响应
    isWait = true;
    $.getJSON("/lp/fetch?lastReceived=" + lastReceived, function (data) {
        if (data == null) return;
        $.each(data, function (i, event) {
            var li = document.createElement('li');

            switch (event.Type) {
            case 0: // JOIN
                if (event.User == $('#uname').text()) {
                    li.innerText = 'You joined the chat room.';
                } else {
                    li.innerText = event.User + ' joined the chat room.';
                }
                break;
            case 1: // LEAVE
                li.innerText = event.User + ' left the chat room.';
                break;
            case 2: // MESSAGE
                var username = document.createElement('strong');
                var content = document.createElement('span');

                username.innerText = event.User;
                content.innerText = event.Content;

                li.appendChild(username);
                li.appendChild(document.createTextNode(': '));
                li.appendChild(content);

                break;
            }

            $('#chatbox li').first().before(li);  //将新的消息的li添加到已有的li前面

            lastReceived = event.Timestamp;
        });
        isWait = false;  //本次响应结束，可进行下次连接请求
    });
}

// Call fetch every 3 seconds
setInterval(fetch, 3000);

fetch();   //同步消息

$(document).ready(function () {

    var postConecnt = function () {
        var uname = $('#uname').text();
        var content = $('#sendbox').val();
        $.post("/lp/post", {
            uname: uname,
            content: content
        });
        $('#sendbox').val("");
    }

    $('#sendbtn').click(function () {
        postConecnt();
    });
});
