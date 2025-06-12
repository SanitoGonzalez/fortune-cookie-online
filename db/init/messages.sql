CREATE TABLE messages (
    id serial primary key,
    created timestamp not null default now(),
    content text not null,
    author varchar(32) not null,
    creator varchar(32) not null
);

INSERT INTO messages (content, author, creator) 
VALUES 
(
    '모니터 앞 3인 이상 집합 금지.',
    '정강산',
    '정강산'
),
(
    '어?',
    '정강산',
    '정강산'
),
(
    '행복은 환경, 운, 머리가 아니라 상황을 바라보는 시각이 결정한다.',
    '루보미르스키 교수',
    '개발자'
),
(
    '평생 삶의 결정적 순간을 찍으려 발버둥쳤으나, 삶의 모든 순간이 결정적인 순간이었다.',
    '앙리 카르티에 브레송',
    '개발자'
),
(
    '인생에서 가장 중요한 것은 행복이 아니라 살아 있는 것이다.',
    '에리히 프롬',
    '개발자'
);