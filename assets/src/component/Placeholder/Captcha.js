import React from "react";
import ContentLoader from "react-content-loader";

const MyLoader = () => (
    <ContentLoader
        height={80}
        width={200}
        speed={2}
        primaryColor="#f3f3f3"
        secondaryColor="#e4e4e4"
    >
        <rect x="4" y="4" rx="7" ry="7" width="392" height="116" />
    </ContentLoader>
);

function captchaPlacholder() {
    return <MyLoader />;
}

export default captchaPlacholder;
