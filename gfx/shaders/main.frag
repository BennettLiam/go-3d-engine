#version 410 core

in vec2 TexCoord;
in vec3 Normal;

out vec4 FragColor;

uniform sampler2D texture1;
uniform vec3 lightDir;
uniform vec3 ambientColor;
uniform float ambientStrength;

void main() {
    vec3 norm = normalize(Normal);
    vec3 lightDirNormalized = normalize(-lightDir);

    // Ambient
    vec3 ambient = ambientStrength * ambientColor;

    // Diffuse
    float diff = max(dot(norm, lightDirNormalized), 0.0);
    vec3 diffuse = diff * vec3(1.0);

    vec4 texColor = texture(texture1, TexCoord);
    vec3 result = (ambient + diffuse) * texColor.rgb;

    FragColor = vec4(result, texColor.a);
}
